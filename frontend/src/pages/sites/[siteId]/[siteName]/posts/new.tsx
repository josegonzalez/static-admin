import { PostForm } from "@/components/posts/post-form";
import { DEFAULT_FIELDS } from "@/components/templates/template-form";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ToastAction } from "@/components/ui/toast";
import { Toaster } from "@/components/ui/toaster";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getTemplate, savePost } from "@/lib/api";
import { Post } from "@/types/post";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

const DEFAULT_POST: Post = {
  id: "",
  path: "",
  frontmatter: DEFAULT_FIELDS,
  blocks: [],
};

export default function NewPostPage() {
  const { toast } = useToast();
  const router = useRouter();
  const { siteId, siteName, templateId } = router.query;
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [post, setPost] = useState<Post>(DEFAULT_POST);

  useEffect(() => {
    const fetchTemplate = async () => {
      try {
        if (!templateId || templateId === "0") {
          setPost(DEFAULT_POST);
          return;
        }

        const template = await getTemplate(templateId as string);
        const dateField = template.fields.find(
          (field) => field.name === "date",
        );
        if (dateField) {
          dateField.dateTimeValue = "0001-01-01T00:00:00Z";
        }

        setPost({
          ...DEFAULT_POST,
          frontmatter: template.fields,
        });
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch template",
        );
      } finally {
        setIsLoading(false);
      }
    };

    if (siteId && siteName) {
      fetchTemplate();
    }
  }, [siteId, siteName, templateId]);

  const handleSubmit = async (newPost: Post) => {
    if (typeof siteId !== "string") return;

    // Use a temporary ID for new posts
    const tempId = "new";
    const response = await savePost(siteId, tempId, newPost);
    toast({
      title: "Success",
      description: `Post created successfully.`,
      variant: "default",
      action: (
        <ToastAction altText="View PR" asChild>
          <Link href={response.pr_url} target="_blank">
            View PR
          </Link>
        </ToastAction>
      ),
    });

    // Redirect back to posts list
    router.push(`/sites/${siteId}/${siteName}/posts`);
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
      <h1 className="text-3xl font-bold mb-0 space-y-0">New Post</h1>
      <p className="text-xs mt-0 pt-0 space-y-0 text-gray-500">
        Template: {templateId === "0" ? "None" : templateId}
      </p>
      <hr className="my-4" />
      <div className="space-y-6">
        <PostForm
          post={post}
          onSubmit={handleSubmit}
          submitButtonText="Create Post"
        />
      </div>
      <Toaster />
    </>
  );
}

NewPostPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
