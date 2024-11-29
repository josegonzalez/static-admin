import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import MultipleSelector from "@/components/ui/multiple-selector";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getPost } from "@/lib/api";
import { FrontmatterField } from "@/types/frontmatter";
import { Post } from "@/types/post";
import { OutputBlockData } from "@editorjs/editorjs";
import dynamic from "next/dynamic";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
import { FieldValues, useForm } from "react-hook-form";
import * as z from "zod";

const DEFAULT_POST: Post = {
  id: "",
  path: "",
  frontmatter: [],
  blocks: [],
};

const formSchema = z.object({
  path: z.string().min(1, "Path is required"),
  frontmatter: z.record(z.unknown()),
  blocks: z.array(z.any()),
});

export default function EditPostPage() {
  const router = useRouter();
  const { siteId, siteName, postId } = router.query;
  const [isLoading, setIsLoading] = useState(true);
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

  const {
    register,
    handleSubmit: handleSubmitForm,
    setValue,
    formState: { errors },
  } = useForm();

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

        let initialStringSliceValues: Record<string, string[]> = {};
        for (const key in postData.frontmatter) {
          const field = postData.frontmatter[key];
          if (field.type === "stringSlice") {
            let values = [];
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

  const handleSubmit = (data: FieldValues, e?: React.BaseSyntheticEvent) => {
    e?.preventDefault();

    setBlocks(data.blocks);
    setValue("blocks", data.blocks);
    setInitialBlockState(data.blocks);

    const newPost: Post = {
      id: post.id,
      path: post.path,
      blocks: data.blocks,
      frontmatter: [],
    };

    for (const key in post.frontmatter) {
      let field: FrontmatterField = {
        name: key,
        type: post.frontmatter[key].type,
        stringValue: "",
        boolValue: false,
        numberValue: 0,
        dateTimeValue: "",
        stringSliceValue: [],
      };
      if (post.frontmatter[key].type === "string") {
        field.stringValue = data[key];
      } else if (post.frontmatter[key].type === "bool") {
        field.boolValue = data[key];
      } else if (post.frontmatter[key].type === "number") {
        field.numberValue = data[key];
      } else if (post.frontmatter[key].type === "dateTime") {
        field.dateTimeValue = data[key];
      } else if (post.frontmatter[key].type === "stringSlice") {
        field.stringSliceValue = data[key];
        let values = [];
        for (const value of data[key]) {
          values.push(value.toString());
        }
        setStringSliceValues({
          ...stringSliceValues,
          [key]: values,
        });
      } else {
        throw new Error(
          `Unknown frontmatter type: ${post.frontmatter[key].type}`,
        );
      }
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
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Edit Post</h1>
      </div>

      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-semibold">Post ID</h2>
          <p className="font-mono">{post.id}</p>
        </div>

        <div>
          <h2 className="text-lg font-semibold">Site</h2>
          <p>{siteName}</p>
        </div>

        <form onSubmit={handleSubmitForm(handleSubmit)} className="space-y-8">
          <div className="space-y-4">
            {Object.values(post.frontmatter).map((field) => {
              return (
                <div key={field.name}>
                  {field.type !== "bool" && (
                    <label htmlFor={field.name} className="mb-2 flex flex-col">
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
                          let values = [];
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
  );
}

EditPostPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};