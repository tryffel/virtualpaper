import {
  ChipField,
  Datagrid,
  ListContextProvider,
  NumberField,
  RaRecord,
  useCreatePath,
  useGetOne,
  TextField as RaTextField,
  Identifier,
} from "react-admin";
import { Grid, Typography } from "@mui/material";
import TextField from "@mui/material/TextField";
import { ChangeEvent } from "react";

export const SearchMetadataResult = ({ query }: SearchProps) => {
  const { data, isLoading } = useGetOne<Result>(
    "metadata/search",
    {
      id: "",
      meta: {
        filter: { q: query },
      },
    },
    { enabled: !!query },
  );

  return (
    <Grid container spacing={1}>
      <Grid item xs={12} sx={{ ml: 1, mr: 1, mb: 2 }}>
        <Typography variant={"h5"}>Keys</Typography>
      </Grid>
      <Grid item xs={12}>
        <MetadataKeyList result={data} loading={isLoading} />
      </Grid>
      <Grid item xs={12} sx={{ margin: "10px 20px" }}>
        <Typography variant={"h5"}>Values</Typography>
      </Grid>
      <Grid item xs={12}>
        <MetadataValueList result={data} loading={isLoading} />
      </Grid>
    </Grid>
  );
};

export type SearchProps = {
  query: string;
  setQuery: (query: string) => void;
};

type Result = {
  id: string;
  keys: object[];
  values: object[];
};

export const SearchMetadataField = ({ query, setQuery }: SearchProps) => {
  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    setQuery(event.target.value);
  };
  return (
    <TextField
      value={query}
      onChange={handleChange}
      size={"small"}
      placeholder={"Search"}
    />
  );
};

const MetadataKeyList = ({
  result,
  loading,
}: {
  result?: Result;
  loading: boolean;
}) => {
  const ctx = {
    resource: "metadata/keys",
    data: result?.keys ?? [],
    page: 1,
    perPage: 10,
    isLoading: loading,
    setSort: () => {},
    sort: { field: "id", value: "ASC" },
    defaultTitle: "Keys",
    showFilter: false,
    total: 10,
    totalPages: 1,
  };
  return (
    // @ts-ignore
    <ListContextProvider value={ctx}>
      <Datagrid rowClick={"edit"}>
        <ChipField source="key" label={"Name"} sortable={false} />
        <RaTextField source="comment" label={"Description"} sortable={false} />
        <NumberField
          source="metadata_values_count"
          label={"Total keys"}
          sortable={false}
        />
        <NumberField
          source="documents_count"
          label={"Total documents"}
          sortable={false}
        />
      </Datagrid>
    </ListContextProvider>
  );
};

const MetadataValueList = ({
  result,
  loading,
}: {
  result?: Result;
  loading: boolean;
}) => {
  const createPath = useCreatePath();
  const ctx = {
    resource: "metadata/values",
    data: result?.values ?? [],
    page: 1,
    perPage: 10,
    isLoading: loading,
    setSort: () => {},
    sort: { field: "id", value: "ASC" },
    defaultTitle: "Keys",
    showFilter: false,
    total: 10,
    totalPages: 1,
  };

  const rowClick = (_: Identifier, resource: string, record: RaRecord) => {
    const path = createPath({ resource, id: record.key_id, type: "edit" });
    return path;
  };

  return (
    // @ts-ignore
    <ListContextProvider value={ctx}>
      <Datagrid rowClick={rowClick}>
        <ChipField source="key" label={"Key"} sortable={false} />
        <ChipField source="value" label={"Value"} sortable={false} />
        <NumberField
          source="documents_count"
          label={"Total documents"}
          sortable={false}
        />
      </Datagrid>
    </ListContextProvider>
  );
};
