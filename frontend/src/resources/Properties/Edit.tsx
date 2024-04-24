import { Edit } from "react-admin";
import { PropertyForm } from "@resources/Properties/Form.tsx";

export const PropertyEdit = () => {
  return (
    <Edit>
      <PropertyForm />
    </Edit>
  );
};
