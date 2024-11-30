// @ts-ignore
import Delimiter from "@coolbytes/editorjs-delimiter";
// @ts-ignore
import editorjsCodecup from "@calumk/editorjs-codecup";
import InlineCode from "@editorjs/inline-code";

import EditorJS, { BlockMutationEvent } from "@editorjs/editorjs";
import {
  OutputBlockData,
  OutputData,
} from "@editorjs/editorjs/types/data-formats";
import Header from "@editorjs/header";
import ImageTool from "@editorjs/image";
import EditorjsList from "@editorjs/list";
import Quote from "@editorjs/quote";
// @ts-ignore
import RawTool from "@editorjs/raw";
import Table from "@editorjs/table";
// @ts-ignore
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
      editorRef.current = new EditorJS(editorConfig as any);
      isReady.current = true;
    }
  }, []);

  return <div id="editorjs" />;
}

export default EditorComponent;
