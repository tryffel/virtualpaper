import { Labeled, SelectArrayInput } from "react-admin";

export const RuleTriggerInput = () => {
  return (
    <Labeled label={"Run rule after document has been:"}>
      <SelectArrayInput
        source={"triggers"}
        label={"trigger type"}
        defaultValue={"document-create"}
        required
        choices={[
          {
            id: "document-create",
            name: "Created",
          },
          {
            id: "document-update",
            name: "Updated",
          },
        ]}
      />
    </Labeled>
  );
};
