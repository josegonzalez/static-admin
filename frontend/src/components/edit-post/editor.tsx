// @ts-ignore
import Delimiter from "@coolbytes/editorjs-delimiter";
import CodeTool from "@editorjs/code";
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
    tools: {
      alert: Alert,
      code: CodeTool,
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
      header: Header,
      image: ImageTool,
      list: EditorjsList,
      raw: RawTool,
      quote: Quote,
      table: Table,
    },
  };

  // This will run only once
  useEffect(() => {
    if (!isReady.current) {
      editorRef.current = new EditorJS(editorConfig);
      isReady.current = true;
    }
  }, []);

  return <div id="editorjs" />;
}

export default EditorComponent;
