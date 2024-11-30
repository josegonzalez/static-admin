import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import MultipleSelector from "@/components/ui/multiple-selector";
import { Spinner } from "@/components/ui/spinner";
import { Toaster } from "@/components/ui/toaster";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getPost, savePost } from "@/lib/api";
import { FrontmatterField } from "@/types/frontmatter";
import { Post } from "@/types/post";
import { OutputBlockData } from "@editorjs/editorjs";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import { FieldValues, useForm } from "react-hook-form";

const DEFAULT_POST: Post = {
  id: "",
  path: "",
  frontmatter: [],
  blocks: [],
};

export default function EditPostPage() {
  const { toast } = useToast();
  const router = useRouter();
  const { siteId, postId } = router.query;
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [post, setPost] = useState<Post>(DEFAULT_POST);
  const [blocks, setBlocks] = useState<OutputBlockData[]>([]);
  const [initialBlockState, setInitialBlockState] = useState<OutputBlockData[]>(
    [],
  );

  const EditorComponent = dynamic(
    () => import("@/components/edit-post/editor"),
    { ssr: false },
  );

  const { register, handleSubmit: handleSubmitForm, setValue } = useForm();

  const [stringSliceValues, setStringSliceValues] = useState<
    Record<string, string[]>
  >({});

  useEffect(() => {
    const fetchPost = async () => {
      try {
        if (typeof siteId !== "string" || typeof postId !== "string") return;
        const postData = await getPost(siteId, postId);
        setPost(postData);
        setBlocks(postData.blocks);
        setInitialBlockState(postData.blocks);

        const initialStringSliceValues: Record<string, string[]> = {};
        for (const key in postData.frontmatter) {
          const field = postData.frontmatter[key];
          if (field.type === "stringSlice") {
            const values = [];
            for (const value of field.stringSliceValue) {
              values.push(value.toString());
            }
            initialStringSliceValues[field.name] = values;
          }
        }
        setStringSliceValues(initialStringSliceValues);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch post");
      } finally {
        setIsLoading(false);
      }
    };

    if (siteId && postId) {
      fetchPost();
    }
  }, [siteId, postId]);

  const handleSubmit = async (
    data: FieldValues,
    e?: React.BaseSyntheticEvent,
  ) => {
    e?.preventDefault();

    setBlocks(data.blocks);
    setValue("blocks", data.blocks);
    setInitialBlockState(data.blocks);
    setIsSaving(true);

    const newPost: Post = {
      id: post.id,
      path: post.path,
      blocks: data.blocks,
      frontmatter: [],
    };

    for (const key in post.frontmatter) {
      const fieldName = post.frontmatter[key].name;
      const field: FrontmatterField = {
        name: fieldName,
        type: post.frontmatter[key].type,
        stringValue: "",
        boolValue: false,
        numberValue: 0,
        dateTimeValue: "0001-01-01T00:00:00Z",
        stringSliceValue: [],
      };
      if (post.frontmatter[key].type === "string") {
        field.stringValue = data[fieldName];
      } else if (post.frontmatter[key].type === "bool") {
        field.boolValue = data[fieldName];
      } else if (post.frontmatter[key].type === "number") {
        field.numberValue = data[fieldName];
      } else if (post.frontmatter[key].type === "dateTime") {
        field.dateTimeValue = data[fieldName];
      } else if (post.frontmatter[key].type === "stringSlice") {
        field.stringSliceValue = data[fieldName];
        setStringSliceValues({
          ...stringSliceValues,
          [fieldName]: data[fieldName],
        });
      } else {
        throw new Error(
          `Unknown frontmatter type: ${post.frontmatter[key].type}`,
        );
      }
      newPost.frontmatter.push(field);
    }

    try {
      if (typeof siteId !== "string" || typeof postId !== "string") return;
      const response = await savePost(siteId, postId, newPost);
      toast({
        title: "Success",
        description: `Post saved successfully. See <a href="${response.prUrl}" target="_blank">${response.prUrl}</a> for details.`,
        variant: "default",
      });
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

  if (isLoading) return null;

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <>
      <h1 className="text-3xl font-bold mb-0 space-y-0">Edit Post</h1>
      <p className="text-xs mt-0 pt-0 space-y-0 text-gray-500">id: {post.id}</p>
      <p className="text-xs mt-0 pt-0 space-y-0 text-gray-500">
        path: {post.path}
      </p>
      {isSaving && (
        <div className="fixed inset-0 z-50 bg-black/80  data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0">
          <Spinner show={true} className="fixed left-[50%] top-[50%] z-50" />
        </div>
      )}
      <hr className="my-4" />
      <div className="space-y-6">
        <div className="space-y-4">
          <form onSubmit={handleSubmitForm(handleSubmit)} className="space-y-8">
            <div className="space-y-4">
              {Object.values(post.frontmatter).map((field) => {
                return (
                  <div key={field.name}>
                    {field.type !== "bool" && (
                      <label
                        htmlFor={field.name}
                        className="mb-2 flex flex-col"
                      >
                        {field.name}
                      </label>
                    )}
                    {field.type === "string" && (
                      <Input
                        id={field.name}
                        {...register(field.name, {
                          value: field.stringValue,
                        })}
                      />
                    )}
                    {field.type === "bool" && (
                      <>
                        <Checkbox
                          id={field.name}
                          {...register(field.name, {
                            value: field.boolValue,
                          })}
                        />
                        <label htmlFor={field.name} className="mb-2 ml-2">
                          {field.name}
                        </label>
                      </>
                    )}
                    {field.type === "number" && (
                      <Input
                        id={field.name}
                        {...register(field.name, {
                          value: field.numberValue,
                        })}
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
                            const values = [];
                            for (const v of value) {
                              values.push(v.value.toString());
                            }
                            setValue(field.name, values);
                          }}
                        />
                      </>
                    )}
                    {field.type === "dateTime" && (
                      <>
                        <input
                          type="hidden"
                          {...register(field.name, {
                            value: field.dateTimeValue,
                          })}
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
                );
              })}
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
            <Button type="submit">Save changes</Button>
          </form>
        </div>
      </div>
      <Toaster />
    </>
  );
}

EditPostPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
