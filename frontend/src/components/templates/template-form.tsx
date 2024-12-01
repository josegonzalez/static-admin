import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { FrontmatterField } from "@/types/frontmatter";
import { DndContext, DragEndEvent, closestCenter } from "@dnd-kit/core";
import { restrictToVerticalAxis } from "@dnd-kit/modifiers";
import {
  SortableContext,
  arrayMove,
  useSortable,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { GripVertical } from "lucide-react";
import { memo, useState } from "react";

export const FIELD_TYPES = [
  { value: "string", label: "Text" },
  { value: "bool", label: "Boolean" },
  { value: "number", label: "Number" },
  { value: "dateTime", label: "Date & Time" },
  { value: "stringSlice", label: "Tags/List" },
];

const CurrentDate = new Date().toISOString();

export const DEFAULT_FIELDS: FrontmatterField[] = [
  {
    name: "title",
    type: "string",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "0001-01-01T00:00:00Z",
    stringSliceValue: [],
  },
  {
    name: "date",
    type: "dateTime",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: CurrentDate,
    stringSliceValue: [],
  },
  {
    name: "category",
    type: "string",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "0001-01-01T00:00:00Z",
    stringSliceValue: [],
  },
  {
    name: "tags",
    type: "stringSlice",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "0001-01-01T00:00:00Z",
    stringSliceValue: [],
  },
  {
    name: "permalink",
    type: "string",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "0001-01-01T00:00:00Z",
    stringSliceValue: [],
  },
  {
    name: "published",
    type: "bool",
    stringValue: "",
    boolValue: false,
    numberValue: 0,
    dateTimeValue: "0001-01-01T00:00:00Z",
    stringSliceValue: [],
  },
];

interface DefaultValueInputProps {
  type: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  value: any;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  onChange: (value: any) => void;
  disabled?: boolean;
}

function DefaultValueInput({
  type,
  value,
  onChange,
  disabled,
}: DefaultValueInputProps) {
  switch (type) {
    case "string":
      return (
        <Input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder="Default text"
          disabled={disabled}
        />
      );
    case "bool":
      return (
        <Checkbox
          checked={value}
          onCheckedChange={onChange}
          disabled={disabled}
        />
      );
    case "number":
      return (
        <Input
          type="number"
          value={value}
          onChange={(e) => onChange(Number(e.target.value))}
          placeholder="Default number"
          disabled={disabled}
        />
      );
    case "dateTime":
      return (
        <DateTimePicker
          value={value ? new Date(value) : undefined}
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          onChange={(date: any) => onChange(date?.toISOString())}
          disabled={disabled}
        />
      );
    case "stringSlice":
      return (
        <MultipleSelector
          value={(value || []).map((v: string) => ({ value: v, label: v }))}
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          onChange={(selected: any) =>
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
            onChange(selected.map((s: any) => s.value))
          }
          creatable
          disabled={disabled}
        />
      );
    default:
      return null;
  }
}

interface NewFieldRowProps {
  onAdd: (
    name: string,
    type: string,
    value: string | number | boolean | string[],
  ) => boolean;
  isAdding: boolean;
  setIsAdding: (isAdding: boolean) => void;
}

function NewFieldRow({ onAdd, isAdding, setIsAdding }: NewFieldRowProps) {
  const [name, setName] = useState("");
  const [type, setType] = useState("");
  const [defaultValue, setDefaultValue] = useState<
    string | number | boolean | string[]
  >("");

  const handleSave = () => {
    if (!name || !type) return;
    if (onAdd(name, type, defaultValue)) {
      setName("");
      setType("");
      setDefaultValue("");
      setIsAdding(false);
    }
  };

  if (!isAdding) {
    return (
      <TableRow>
        <TableCell colSpan={4}>
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

const SortableRow = memo(function SortableRow({
  id,
  field,
  isDefaultField,
  onDelete,
  onUpdate,
}: {
  id: string;
  field: FrontmatterField;
  isDefaultField: boolean;
  onDelete: (name: string) => void;
  onUpdate: (oldName: string, field: FrontmatterField) => void;
}) {
  const [isEditing, setIsEditing] = useState(false);
  const [name, setName] = useState(field.name);
  const [type, setType] = useState(field.type);
  const [value, setValue] = useState<string | number | boolean | string[]>(
    () => {
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
    },
  );

  const handleSave = () => {
    const updatedField: FrontmatterField = {
      name,
      type,
      stringValue: type === "string" ? String(value) : "",
      boolValue: type === "bool" ? Boolean(value) : false,
      numberValue: type === "number" ? Number(value) : 0,
      dateTimeValue: type === "dateTime" ? String(value) : "",
      stringSliceValue: type === "stringSlice" ? (value as string[]) : [],
    };
    onUpdate(field.name, updatedField);
    setIsEditing(false);
  };

  const handleCancel = () => {
    setName(field.name);
    setType(field.type);
    setValue(() => {
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
    setIsEditing(false);
  };

  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <TableRow ref={setNodeRef} style={style}>
      <TableCell>
        <div className="flex items-center gap-2">
          {!isDefaultField && (
            <Button
              variant="ghost"
              {...attributes}
              {...listeners}
              className="p-0 cursor-grab"
            >
              <GripVertical className="h-4 w-4" />
            </Button>
          )}
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
          FIELD_TYPES.find((t) => t.value === field.type)?.label
        )}
      </TableCell>
      <TableCell>
        {isEditing ? (
          <DefaultValueInput
            type={field.type}
            value={value}
            onChange={setValue}
            disabled={isDefaultField}
          />
        ) : (
          <DefaultValueInput
            type={field.type}
            value={value}
            onChange={() => {}}
            disabled={true}
          />
        )}
      </TableCell>
      <TableCell>
        {isDefaultField ? (
          <span className="text-sm text-muted-foreground">Default field</span>
        ) : isEditing ? (
          <div className="flex gap-2">
            <Button variant="ghost" size="sm" onClick={handleSave}>
              Save
            </Button>
            <Button variant="ghost" size="sm" onClick={handleCancel}>
              Cancel
            </Button>
          </div>
        ) : (
          <div className="flex gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsEditing(true)}
            >
              Edit
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="text-red-500 hover:text-red-700"
              onClick={() => onDelete(field.name)}
            >
              Delete
            </Button>
          </div>
        )}
      </TableCell>
    </TableRow>
  );
});

interface TemplateFormProps {
  templateName: string;
  setTemplateName: (name: string) => void;
  fields: FrontmatterField[];
  setFields: React.Dispatch<React.SetStateAction<FrontmatterField[]>>;
  error: string | null;
  setError: (error: string | null) => void;
  isAddingField: boolean;
  setIsAddingField: (isAdding: boolean) => void;
  isSaving: boolean;
  onSave: () => void;
  onCancel: () => void;
  saveButtonText: string;
}

export function TemplateForm({
  templateName,
  setTemplateName,
  fields,
  setFields,
  error,
  setError,
  isAddingField,
  setIsAddingField,
  isSaving,
  onSave,
  onCancel,
  saveButtonText,
}: TemplateFormProps) {
  const handleAddField = (
    name: string,
    type: string,
    value: string | number | boolean | string[],
  ) => {
    if (fields.some((f) => f.name === name)) {
      setError("A field with this name already exists");
      return false;
    }

    const field: FrontmatterField = {
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
        if (typeof value === "string") {
          field.stringValue = value;
        }
        break;
      case "bool":
        if (typeof value === "boolean") {
          field.boolValue = value;
        }
        break;
      case "number":
        if (typeof value === "number") {
          field.numberValue = value;
        }
        break;
      case "dateTime":
        if (typeof value === "string") {
          field.dateTimeValue = value;
        }
        break;
      case "stringSlice":
        if (Array.isArray(value)) {
          field.stringSliceValue = value;
        }
        break;
      default:
        throw new Error(`Unknown field type: ${type}`);
    }

    setFields([...fields, field]);
    setError(null);
    return true;
  };

  const handleDelete = (name: string) => {
    if (DEFAULT_FIELDS.some((f) => f.name === name)) {
      return;
    }

    setFields((currentFields: FrontmatterField[]) =>
      currentFields.filter((field: FrontmatterField) => field.name !== name),
    );
  };

  const handleUpdate = (oldName: string, updatedField: FrontmatterField) => {
    setFields((currentFields: FrontmatterField[]) =>
      currentFields.map((field: FrontmatterField) =>
        field.name === oldName ? updatedField : field,
      ),
    );
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over) return;

    if (active.id !== over.id) {
      setFields((items: FrontmatterField[]) => {
        const oldIndex = items.findIndex(
          (item: FrontmatterField) => item.name === active.id,
        );
        const newIndex = items.findIndex(
          (item: FrontmatterField) => item.name === over.id,
        );
        return arrayMove(items, oldIndex, newIndex);
      });
    }
  };

  return (
    <div className="space-y-4">
      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div>
        <Label htmlFor="templateName">Template Name</Label>
        <Input
          id="templateName"
          value={templateName}
          onChange={(e) => setTemplateName(e.target.value)}
          placeholder="Enter template name"
          className="max-w-md"
          required
        />
      </div>

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
              modifiers={[restrictToVerticalAxis]}
            >
              <SortableContext
                items={fields.map((f) => f.name)}
                strategy={verticalListSortingStrategy}
              >
                {fields.map((field) => {
                  const isDefaultField = DEFAULT_FIELDS.some(
                    (f) => f.name === field.name,
                  );
                  return (
                    <SortableRow
                      key={field.name}
                      id={field.name}
                      field={field}
                      isDefaultField={isDefaultField}
                      onDelete={handleDelete}
                      onUpdate={handleUpdate}
                    />
                  );
                })}
              </SortableContext>
            </DndContext>
            <NewFieldRow
              onAdd={handleAddField}
              isAdding={isAddingField}
              setIsAdding={setIsAddingField}
            />
          </TableBody>
        </Table>
      </div>

      <div className="flex justify-end gap-4">
        <Button variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button onClick={onSave} disabled={isSaving || isAddingField}>
          {isSaving ? "Saving..." : saveButtonText}
        </Button>
      </div>
    </div>
  );
}
