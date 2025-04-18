import { PostForm } from "@/components/posts/post-form";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { ToastAction } from "@/components/ui/toast";
import { Toaster } from "@/components/ui/toaster";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getPost, savePost } from "@/lib/api";
import { Post } from "@/types/post";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

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
  const [error, setError] = useState<string | null>(null);
  const [post, setPost] = useState<Post>(DEFAULT_POST);

  useEffect(() => {
    const fetchPost = async () => {
      try {
        if (typeof siteId !== "string" || typeof postId !== "string") return;
        const postData = await getPost(siteId, postId);
        setPost(postData);
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

  const handleSubmit = async (newPost: Post) => {
    if (typeof siteId !== "string" || typeof postId !== "string") return;
    const response = await savePost(siteId, postId, newPost);
    toast({
      title: "Success",
      description: `Post saved successfully.`,
      variant: "default",
      action: (
        <ToastAction altText="View PR" asChild>
          <Link href={response.pr_url} target="_blank">
            View PR
          </Link>
        </ToastAction>
      ),
    });
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
      <hr className="my-4" />
      <div className="space-y-6">
        <PostForm post={post} onSubmit={handleSubmit} />
      </div>
      <Toaster />
    </>
  );
}

EditPostPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
