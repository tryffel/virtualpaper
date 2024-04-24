import { Create } from "react-admin";
import { PropertyForm } from "@resources/Properties/Form.tsx";

export const PropertyCreate = () => {
  return (
    <Create>
      <PropertyForm />
    </Create>
  );
};
