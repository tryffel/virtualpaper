import { SelectInput } from "react-admin";

export const IconColorSelect = () => {
  const choices = [
    { id: "primary", name: "primary" },
    { id: "secondary", name: "secondary" },
    { id: "info", name: "info" },
    { id: "success", name: "success" },
    { id: "warning", name: "warning" },
    { id: "error", name: "error" },
  ];

  return (
    <SelectInput choices={choices} emptyValue={""} source={"style.color"} />
  );
};
