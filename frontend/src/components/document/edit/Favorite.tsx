import { IconButton } from "@mui/material";
import BookmarkIcon from "@mui/icons-material/Bookmark";
import { useInput } from "ra-core";
import BookmarkBorderIcon from "@mui/icons-material/BookmarkBorder";

export const FavoriteDocumentInput = ({ source }: { source: string }) => {
  const { field } = useInput({ source });

  const handleChange = () => {
    console.log("value", field.value);
    field.onChange(!field.value);
  };

  return (
    <IconButton onClick={handleChange}>
      {field.value ? (
        // @ts-ignore
        <BookmarkIcon color={"favorite"} />
      ) : (
        <BookmarkBorderIcon />
      )}
    </IconButton>
  );
};
