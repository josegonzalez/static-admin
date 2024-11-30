// @ts-expect-error: editorjs-delimiter does not use typescript
import Delimiter from "@coolbytes/editorjs-delimiter";
// @ts-expect-error: editorjs-codecup does not use typescript
import editorjsCodecup from "@calumk/editorjs-codecup";
import InlineCode from "@editorjs/inline-code";

import EditorJS, { BlockMutationEvent, EditorConfig } from "@editorjs/editorjs";
import {
  OutputBlockData,
  OutputData,
} from "@editorjs/editorjs/types/data-formats";
import Header from "@editorjs/header";
import ImageTool from "@editorjs/image";
import EditorjsList from "@editorjs/list";
import Quote from "@editorjs/quote";
// @ts-expect-error: editorjs-raw does not use typescript
import RawTool from "@editorjs/raw";
import Table from "@editorjs/table";
// @ts-expect-error: editorjs-alert does not use typescript
import Alert from "editorjs-alert";
import { useEffect, useRef } from "react";

interface EditorProps {
  blocks: OutputBlockData[];
  onChange: (blocks: OutputBlockData[]) => void;
}

export function EditorComponent({ blocks, onChange }: EditorProps) {
  const isReady = useRef(false);

  const editorRef = useRef<EditorJS | null>(null);

  const editorConfig = {
    holder: "editorjs",
    data: {
      blocks: blocks,
    },
    minHeight: 100,
    onChange: async (
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      _: any,
      event: BlockMutationEvent | BlockMutationEvent[],
    ) => {
      if (Array.isArray(event)) {
        event.forEach((e) => {
          e.preventDefault();
        });
      } else {
        event.preventDefault();
      }
      if (editorRef.current) {
        editorRef.current.saver
          .save()
          .then((outputData: OutputData) => {
            onChange(outputData.blocks);
          })
          .catch((err) => {
            console.error(err);
          });
      }
    },
    placeholder: "Write something...",
    tools: {
      alert: Alert,
      code: {
        class: editorjsCodecup,
        shortcut: "CMD+SHIFT+D",
      },
      delimiter: {
        class: Delimiter,
        config: {
          defaultLineWidth: 100,
          defaultStyle: "line",
          lineWidthOptions: [100],
          lineThicknessOptions: [2],
          styleOptions: ["line"],
        },
      },
      inlineCode: {
        class: InlineCode,
        shortcut: "CMD+SHIFT+C",
      },
      header: {
        class: Header,
        inlineToolbar: true,
      },
      image: {
        class: ImageTool,
        inlineToolbar: true,
      },
      list: {
        class: EditorjsList,
        inlineToolbar: true,
      },
      raw: RawTool,
      quote: {
        class: Quote,
        inlineToolbar: true,
      },
      table: {
        class: Table,
        inlineToolbar: true,
      },
    },
  };

  // This will run only once
  useEffect(() => {
    if (!isReady.current) {
      // @ts-expect-error: certain tools are not typed
      editorRef.current = new EditorJS(editorConfig as EditorConfig);
      isReady.current = true;
    }
  });

  return <div id="editorjs" />;
}

export default EditorComponent;
