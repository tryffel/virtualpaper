import {
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Typography,
  List,
  IconButton,
} from "@mui/material";
import {
  Button,
  Form,
  RaRecord,
  RecordContextProvider,
  TextField,
  useGetList,
  useNotify,
  useRecordContext,
  useRefresh,
  useUpdate,
} from "react-admin";
import ListItem from "@mui/material/ListItem";
import CancelIcon from "@mui/icons-material/Cancel";
import SaveIcon from "@mui/icons-material/Save";
import DragHandleIcon from "@mui/icons-material/DragHandle";
import * as React from "react";
import get from "lodash/get";
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  MouseSensor,
  TouchSensor,
  useSensor,
  useSensors,
} from "@dnd-kit/core";
import {
  arrayMove,
  SortableContext,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import ListItemText from "@mui/material/ListItemText";

export interface ReorderProps {
  modalOpen: boolean;
  setModalOpen: (open: boolean) => void;
}

export const ReorderRulesDialog = (props: ReorderProps) => {
  const { modalOpen, setModalOpen } = props;
  const updateLoading = false;

  const [originalIds, setOriginalIds] = React.useState<Array<number>>([]);
  const [updatedIds, setUpdatedIds] = React.useState<Array<number>>([]);
  const [activeId, setActiveId] = React.useState<number | null>(null);
  const notify = useNotify();
  const refresh = useRefresh();

  const mouseSensor = useSensor(MouseSensor);
  const touchSensor = useSensor(TouchSensor);
  const keyboardSensor = useSensor(KeyboardSensor);

  const sensors = useSensors(mouseSensor, touchSensor, keyboardSensor);

  const { data } = useGetList(
    "processing/rules",
    {
      pagination: { page: 1, perPage: 500 },
      sort: { field: "order", order: "DESC" },
    },
    {
      onSuccess: (data) => {
        const ids = data.data.map((item) => get(item, "id"));
        setOriginalIds(ids);
        setUpdatedIds(ids);
      },
    },
  );

  const [update] = useUpdate(
    "reorder-rules",
    {},
    {
      onSuccess: () => {
        notify("Order saved");
        setModalOpen(false);
        refresh();
      },
      onError: (data) => {
        notify(`${data}`, { type: "error" });
      },
    },
  );

  const handleSave = () => {
    update("reorder-rules", {
      id: "rules",
      data: { ids: updatedIds },
      previousData: { ids: originalIds },
    });
  };
  const handleCancel = () => {
    setModalOpen(false);
  };

  const handleDragEvent = (event: any) => {
    const { active, over } = event;
    if (active.id != over.id) {
      const oldIndex = updatedIds.indexOf(Number(active.id));
      const newIndex = updatedIds.indexOf(Number(over.id));
      const newArray = arrayMove(updatedIds, oldIndex, newIndex);
      setUpdatedIds(newArray);
    }
    setActiveId(null);
  };

  const handleDragStart = (event: any) => {
    setActiveId(event.active.id);
  };

  return (
    <Dialog open={modalOpen}>
      <DialogTitle>Reorder rules</DialogTitle>
      <Form>
        <DialogContent>
          <Typography variant="body1" sx={{ pb: 2 }}>
            Reorder rules by dragging and dropping them
          </Typography>
          {data && (
            <List>
              <DndContext
                sensors={sensors}
                onDragEnd={handleDragEvent}
                onDragStart={handleDragStart}
              >
                <SortableContext
                  items={data}
                  strategy={verticalListSortingStrategy}
                >
                  {updatedIds.map((id) => (
                    <RuleEntry
                      key={id}
                      record={data?.find((entry) => get(entry, "id") == id)}
                    />
                  ))}
                </SortableContext>
                <DragOverlay>
                  {activeId && (
                    <RuleEntry
                      key={activeId}
                      record={data?.find(
                        (entry) => get(entry, "id") == activeId,
                      )}
                    />
                  )}
                </DragOverlay>
              </DndContext>
            </List>
          )}
        </DialogContent>

        <DialogActions>
          <Button label={"Cancel"} onClick={handleCancel} color={"secondary"}>
            <CancelIcon />
          </Button>

          <Button
            label={updateLoading ? "Saving..." : "Save"}
            onClick={handleSave}
          >
            <SaveIcon />
          </Button>
        </DialogActions>
      </Form>
    </Dialog>
  );
};

const RuleEntry = (props: { record: RaRecord }) => {
  const { record } = props;
  const { attributes, listeners, setNodeRef, transform, transition } =
    useSortable({ id: `${get(record, "id")}` });
  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0)`,
        transition,
      }
    : undefined;

  const text = get(record, "name");
  const disabled = !get(record, "enabled");

  return (
    <RecordContextProvider value={record}>
      <ListItem ref={setNodeRef} style={style}>
        <ListItemText primary={text} secondary={disabled ? "Disabled" : ""} />
        <IconButton
          {...listeners}
          {...attributes}
          edge={"end"}
          style={{ touchAction: "pan-y" }}
        >
          <DragHandleIcon />
        </IconButton>
      </ListItem>
    </RecordContextProvider>
  );
};

export const RuleTitle = (props: object = {}) => {
  const record = useRecordContext(props);
  if (!record) {
    return null;
  }

  const enabled = get(record, "enabled");
  return (
    <TextField sx={{ fontWeight: enabled ? "500" : "50" }} source="name" />
  );
};
