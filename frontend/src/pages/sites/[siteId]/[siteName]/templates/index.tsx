import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import MultipleSelector from "@/components/ui/multiple-selector";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import DashboardLayout from "@/layouts/dashboard-layout";
import { FrontmatterField } from "@/types/frontmatter";
import { DndContext, DragEndEvent, closestCenter } from "@dnd-kit/core";
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
import { useState } from "react";

const FIELD_TYPES = [
  { value: "string", label: "Text" },
  { value: "bool", label: "Boolean" },
  { value: "number", label: "Number" },
  { value: "dateTime", label: "Date & Time" },
  { value: "stringSlice", label: "Tags/List" },
];

const DEFAULT_FIELDS: FrontmatterField[] = [
  {
    name: "title",
    type: "string",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "",
    stringSliceValue: [],
  },
  {
    name: "date",
    type: "dateTime",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "",
    stringSliceValue: [],
  },
];

interface EditFieldModalProps {
  field: FrontmatterField;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (name: string, type: string) => void;
}

function EditFieldModal({
  field,
  open,
  onOpenChange,
  onSave,
}: EditFieldModalProps) {
  const [name, setName] = useState(field.name);
  const [type, setType] = useState(field.type);
  const [error, setError] = useState<string | null>(null);

  const handleSave = () => {
    if (!name || !type) {
      setError("Name and type are required");
      return;
    }
    onSave(name, type);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Field</DialogTitle>
          <DialogDescription>Edit the field's name and type.</DialogDescription>
        </DialogHeader>
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium">Field Name</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter field name"
              disabled={DEFAULT_FIELDS.some((f) => f.name === field.name)}
            />
          </div>
          <div>
            <label className="text-sm font-medium">Field Type</label>
            <Select
              value={type}
              onValueChange={setType}
              disabled={DEFAULT_FIELDS.some((f) => f.name === field.name)}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select field type" />
              </SelectTrigger>
              <SelectContent>
                {FIELD_TYPES.map((type) => (
                  <SelectItem key={type.value} value={type.value}>
                    {type.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave}>Save Changes</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

interface EditableRowProps {
  id: string;
  field: FrontmatterField;
  onDelete: (name: string) => void;
  onSave: (name: string, type: string) => void;
}

interface DefaultValueInputProps {
  type: string;
  value: any;
  onChange: (value: any) => void;
}

function DefaultValueInput({ type, value, onChange }: DefaultValueInputProps) {
  switch (type) {
    case "string":
      return (
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Default text"
        />
      );
    case "bool":
      return <Checkbox checked={value} onCheckedChange={onChange} />;
    case "number":
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          placeholder="Default number"
        />
      );
    case "dateTime":
      return (
        <DateTimePicker
          value={value ? new Date(value) : undefined}
          onChange={(date: any) => onChange(date?.toISOString())}
        />
      );
    case "stringSlice":
      return (
        <MultipleSelector
          value={(value || []).map((v: string) => ({ value: v, label: v }))}
          onChange={(selected: any) =>
            onChange(selected.map((s: any) => s.value))
          }
          creatable
        />
      );
    default:
      return null;
  }
}

function EditableRow({ id, field, onDelete, onSave }: EditableRowProps) {
  const { attributes, listeners, transform, transition, setNodeRef } =
    useSortable({ id });
  const [isEditing, setIsEditing] = useState(false);
  const [name, setName] = useState(field.name);
  const [type, setType] = useState(field.type);
  const [defaultValue, setDefaultValue] = useState(() => {
    switch (field.type) {
      case "string":
        return field.stringValue;
      case "bool":
        return field.boolValue;
      case "number":
        return field.numberValue;
      case "dateTime":
        return field.dateTimeValue;
      case "stringSlice":
        return field.stringSliceValue;
      default:
        return "";
    }
  });

  const isDefaultField = DEFAULT_FIELDS.some((f) => f.name === field.name);

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const handleSave = () => {
    if (!name || !type) return;
    const updatedField = {
      name,
      type,
      stringValue: type === "string" ? defaultValue : "",
      boolValue: type === "bool" ? defaultValue : false,
      numberValue: type === "number" ? defaultValue : 0,
      dateTimeValue: type === "dateTime" ? defaultValue : "",
      stringSliceValue: type === "stringSlice" ? defaultValue : [],
    };
    onSave(name, type);
    setIsEditing(false);
  };

  const handleCancel = () => {
    setName(field.name);
    setType(field.type);
    setIsEditing(false);
  };

  return (
    <TableRow ref={setNodeRef} style={style}>
      <TableCell>
        <div className="flex items-center gap-2">
          <button
            {...attributes}
            {...listeners}
            className="cursor-grab hover:text-gray-700 active:cursor-grabbing"
          >
            <GripVertical className="h-4 w-4" />
          </button>
          {isEditing ? (
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-[200px]"
              disabled={isDefaultField}
            />
          ) : (
            field.name
          )}
        </div>
      </TableCell>
      <TableCell>
        {isEditing ? (
          <Select
            value={type}
            onValueChange={setType}
            disabled={isDefaultField}
          >
            <SelectTrigger className="w-[200px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {FIELD_TYPES.map((type) => (
                <SelectItem key={type.value} value={type.value}>
                  {type.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          FIELD_TYPES.find((t) => t.value === field.type)?.label || field.type
        )}
      </TableCell>
      <TableCell>
        {isEditing ? (
          <DefaultValueInput
            type={type}
            value={defaultValue}
            onChange={setDefaultValue}
          />
        ) : (
          <div className="text-sm text-gray-500">
            {type === "stringSlice"
              ? (defaultValue as string[]).join(", ")
              : String(defaultValue)}
          </div>
        )}
      </TableCell>
      <TableCell>
        <div className="flex gap-2">
          {isEditing ? (
            <>
              <Button
                variant="ghost"
                size="sm"
                onClick={handleSave}
                disabled={!name || !type}
              >
                Save
              </Button>
              <Button variant="ghost" size="sm" onClick={handleCancel}>
                Cancel
              </Button>
            </>
          ) : (
            <>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsEditing(true)}
                disabled={isDefaultField}
              >
                Edit
              </Button>
              <Button
                variant="ghost"
                size="sm"
                className="text-red-500 hover:text-red-700"
                onClick={() => onDelete(field.name)}
                disabled={isDefaultField}
              >
                Delete
              </Button>
            </>
          )}
        </div>
      </TableCell>
    </TableRow>
  );
}

type NewFieldRowProps = {
  onAdd: (name: string, type: string, value: any) => boolean;
};

function NewFieldRow({ onAdd }: NewFieldRowProps) {
  const [name, setName] = useState("");
  const [type, setType] = useState("");
  const [isAdding, setIsAdding] = useState(false);
  const [defaultValue, setDefaultValue] = useState<any>("");

  const handleSave = () => {
    if (!name || !type) return;
    const field: FrontmatterField = {
      name,
      type,
      stringValue: type === "string" ? defaultValue : "",
      boolValue: type === "bool" ? defaultValue : false,
      numberValue: type === "number" ? defaultValue : 0,
      dateTimeValue: type === "dateTime" ? defaultValue : "",
      stringSliceValue: type === "stringSlice" ? defaultValue : [],
    };
    if (onAdd(field.name, field.type, defaultValue)) {
      setName("");
      setType("");
      setDefaultValue("");
      setIsAdding(false);
    }
  };

  if (!isAdding) {
    return (
      <TableRow>
        <TableCell colSpan={3}>
          <Button
            variant="ghost"
            size="sm"
            className="w-full text-blue-500 hover:text-blue-700"
            onClick={() => setIsAdding(true)}
          >
            + Add new field
          </Button>
        </TableCell>
      </TableRow>
    );
  }

  return (
    <TableRow>
      <TableCell>
        <Input
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="w-[200px]"
          placeholder="Field name"
        />
      </TableCell>
      <TableCell>
        <Select value={type} onValueChange={setType}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Select type" />
          </SelectTrigger>
          <SelectContent>
            {FIELD_TYPES.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                {type.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </TableCell>
      <TableCell>
        {type && (
          <DefaultValueInput
            type={type}
            value={defaultValue}
            onChange={setDefaultValue}
          />
        )}
      </TableCell>
      <TableCell>
        <div className="flex gap-2">
          <Button
            variant="ghost"
            size="sm"
            onClick={handleSave}
            disabled={!name || !type}
          >
            Save
          </Button>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setName("");
              setType("");
              setIsAdding(false);
            }}
          >
            Cancel
          </Button>
        </div>
      </TableCell>
    </TableRow>
  );
}

export default function TemplatesPage() {
  const [fields, setFields] = useState<FrontmatterField[]>(DEFAULT_FIELDS);
  const [error, setError] = useState<string | null>(null);

  const handleAddField = (name: string, type: string, value: any) => {
    if (fields.some((f) => f.name === name)) {
      setError("A field with this name already exists");
      return false;
    }

    let field: FrontmatterField = {
      name,
      type,
      stringValue: "",
      boolValue: false,
      numberValue: 0,
      dateTimeValue: "",
      stringSliceValue: [],
    };

    switch (type) {
      case "string":
        field.stringValue = value;
        break;
      case "bool":
        field.boolValue = value;
        break;
      case "number":
        field.numberValue = value;
        break;
      case "dateTime":
        field.dateTimeValue = value;
        break;
      case "stringSlice":
        field.stringSliceValue = value;
        break;
      default:
        throw new Error(`Unknown field type: ${type}`);
        break;
    }

    setFields([...fields, field]);
    setError(null);
    return true;
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;

    if (active.id !== over.id) {
      setFields((items) => {
        const oldIndex = items.findIndex((item) => item.name === active.id);
        const newIndex = items.findIndex((item) => item.name === over.id);
        return arrayMove(items, oldIndex, newIndex);
      });
    }
  };

  const handleDelete = (name: string) => {
    setFields(fields.filter((f) => f.name !== name));
  };

  const handleSaveEdit = (
    oldName: string,
    newName: string,
    newType: string,
  ) => {
    setFields(
      fields.map((field) =>
        field.name === oldName
          ? { ...field, name: newName, type: newType }
          : field,
      ),
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Frontmatter Templates</h1>
        <hr />
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Field Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Default Value</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <DndContext
              collisionDetection={closestCenter}
              onDragEnd={handleDragEnd}
            >
              <SortableContext
                items={fields.map((f) => f.name)}
                strategy={verticalListSortingStrategy}
              >
                {fields.map((field) => (
                  <EditableRow
                    key={field.name}
                    id={field.name}
                    field={field}
                    onDelete={handleDelete}
                    onSave={(name, type) =>
                      handleSaveEdit(field.name, name, type)
                    }
                  />
                ))}
              </SortableContext>
            </DndContext>
            <NewFieldRow onAdd={handleAddField} />
          </TableBody>
        </Table>
      </div>
    </div>
  );
}

TemplatesPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
