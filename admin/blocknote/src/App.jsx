import "@blocknote/core/fonts/inter.css";
import { BlockNoteView } from "@blocknote/mantine";
import "@blocknote/mantine/style.css";
import { useCreateBlockNote } from "@blocknote/react";
import { zh } from "@blocknote/core/locales";
import "./style.css";

export default function App() {
  // Creates a new editor instance.
  const editor = useCreateBlockNote(
    {
      dictionary: zh,
      tables: {
        splitCells: true,
        cellBackgroundColor: true,
        cellTextColor: true,
        headers: true,
      },
    }
  );

  console.log('editor:', editor);
  window.TaoBlog.blocknote = editor;

  // Renders the editor instance using a React component.
  return <BlockNoteView editor={editor} />;
}
