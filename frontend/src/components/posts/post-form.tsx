import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import MultipleSelector from "@/components/ui/multiple-selector";
import { Spinner } from "@/components/ui/spinner";
import { useToast } from "@/hooks/use-toast";
import { cn } from "@/lib/utils";
import { FrontmatterField } from "@/types/frontmatter";
import { Post } from "@/types/post";
import { OutputBlockData } from "@editorjs/editorjs";
import dynamic from "next/dynamic";
import { useState } from "react";
import { FieldValues, useForm } from "react-hook-form";

const EditorComponent = dynamic(() => import("@/components/edit-post/editor"), {
  ssr: false,
});

interface PostFormProps {
  post: Post;
  onSubmit: (post: Post) => Promise<void>;
  submitButtonText?: string;
}

export function PostForm({
  post,
  onSubmit,
  submitButtonText = "Save changes",
}: PostFormProps) {
  const { toast } = useToast();
  const [isSaving, setIsSaving] = useState(false);
  const [blocks] = useState<OutputBlockData[]>(post.blocks);
  const [initialBlockState] = useState<OutputBlockData[]>(post.blocks);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const [stringSliceValues, setStringSliceValues] = useState<
    Record<string, string[]>
  >(() => {
    const values: Record<string, string[]> = {};
    post.frontmatter.forEach((field) => {
      if (field.type === "stringSlice") {
        values[field.name] = field.stringSliceValue;
      }
    });
    return values;
  });

  const { register, handleSubmit: handleSubmitForm, setValue } = useForm();

  const handleSubmit = async (
    data: FieldValues,
    e?: React.BaseSyntheticEvent,
  ) => {
    e?.preventDefault();
    setFieldErrors({});

    const newErrors: Record<string, string> = {};
    const titleField = post.frontmatter.find((field) => field.name === "title");
    if (
      titleField &&
      (!data[titleField.name] || data[titleField.name].trim() === "")
    ) {
      newErrors[titleField.name] = "Title cannot be empty";
    }

    if (Object.keys(newErrors).length > 0) {
      setFieldErrors(newErrors);
      return;
    }

    setIsSaving(true);

    const newPost: Post = {
      id: post.id,
      path: post.path,
      blocks: data.blocks,
      frontmatter: post.frontmatter.map((field) => {
        const newField: FrontmatterField = { ...field };
        if (field.type !== "dateTime") {
          newField.dateTimeValue = "0001-01-01T00:00:00Z";
        }

        if (field.type === "string") {
          newField.stringValue = data[field.name];
        } else if (field.type === "bool") {
          newField.boolValue = data[field.name];
        } else if (field.type === "number") {
          newField.numberValue = data[field.name];
        } else if (field.type === "dateTime") {
          newField.dateTimeValue = data[field.name];
        } else if (field.type === "stringSlice") {
          newField.stringSliceValue = data[field.name];
        }
        return newField;
      }),
    };

    try {
      await onSubmit(newPost);
    } catch (err) {
      toast({
        title: "Error",
        description: err instanceof Error ? err.message : "Failed to save post",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <>
      {isSaving && (
        <div className="fixed inset-0 z-50 bg-black/80 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0">
          <Spinner show={true} className="fixed left-[50%] top-[50%] z-50" />
        </div>
      )}
      <form onSubmit={handleSubmitForm(handleSubmit)} className="space-y-8">
        <div className="space-y-4">
          {post.frontmatter.map((field) => (
            <div key={field.name}>
              {field.type !== "bool" && (
                <Label
                  htmlFor={field.name}
                  className={cn(
                    "mb-2 flex flex-col",
                    fieldErrors[field.name] && "text-destructive",
                  )}
                >
                  {field.name}
                </Label>
              )}
              {field.type === "string" && (
                <div className="space-y-2">
                  <Input
                    id={field.name}
                    {...register(field.name, { value: field.stringValue })}
                    className={cn(
                      fieldErrors[field.name] && "border-destructive",
                    )}
                  />
                  {fieldErrors[field.name] && (
                    <p className="text-sm text-destructive">
                      {fieldErrors[field.name]}
                    </p>
                  )}
                </div>
              )}
              {field.type === "bool" && (
                <>
                  <Checkbox
                    id={field.name}
                    {...register(field.name, { value: field.boolValue })}
                  />
                  <label htmlFor={field.name} className="mb-2 ml-2">
                    {field.name}
                  </label>
                </>
              )}
              {field.type === "number" && (
                <Input
                  id={field.name}
                  {...register(field.name, { value: field.numberValue })}
                />
              )}
              {field.type === "stringSlice" && (
                <>
                  <select
                    {...register(field.name, {
                      value: stringSliceValues[field.name],
                    })}
                    multiple
                    className="hidden"
                  >
                    {stringSliceValues[field.name].map((option) => (
                      <option key={option} value={option}>
                        {option}
                      </option>
                    ))}
                  </select>
                  <MultipleSelector
                    value={stringSliceValues[field.name].map((value) => ({
                      value: value,
                      label: value,
                    }))}
                    creatable
                    onChange={(value) => {
                      const values = value.map((v) => v.value.toString());
                      setValue(field.name, values);
                      setStringSliceValues({
                        ...stringSliceValues,
                        [field.name]: values,
                      });
                    }}
                  />
                </>
              )}
              {field.type === "dateTime" && (
                <>
                  <input
                    type="hidden"
                    {...register(field.name, { value: field.dateTimeValue })}
                  />
                  <DateTimePicker
                    hourCycle={12}
                    value={
                      field.dateTimeValue
                        ? new Date(field.dateTimeValue)
                        : undefined
                    }
                    onChange={(date) => {
                      if (date) {
                        setValue(field.name, date.toISOString());
                      }
                    }}
                  />
                </>
              )}
            </div>
          ))}
        </div>

        <div>
          <label htmlFor="blocks">Content</label>
          <hr className="my-4" />
          <textarea
            id="blocks"
            {...register("blocks", { value: blocks })}
            className="hidden"
          />
          <EditorComponent
            blocks={initialBlockState}
            onChange={(blocks: OutputBlockData[]) => {
              setValue("blocks", blocks);
            }}
          />
        </div>
        <Button type="submit">{submitButtonText}</Button>
      </form>
    </>
  );
}
