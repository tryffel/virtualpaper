/*
 * Virtualpaper is a service to manage users paper documents in virtual format.
 * Copyright (C) 2022  Tero Vierimaa
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */
import React from 'react'
import { Create, SimpleForm, TextInput } from "react-admin";
const IconSelect = React.lazy(() => import('./IconSelect'))

export const MetadataKeyCreate = () => (
  <Create
    title={<CreateTitle />}
    transform={(data: any) => ({
      ...data,
      style: JSON.stringify(data.style),
    })}
  >
    <SimpleForm defaultValues={{ icon: "Label", style: "{}" }}>
      <TextInput source="key" label="Name" />
      <TextInput source="comment" label="Description" />
      <IconSelect source={"icon"} displayIcon={true} />
    </SimpleForm>
  </Create>
);

const CreateTitle = () => {
  return <span>Add metadata key</span>;
};
