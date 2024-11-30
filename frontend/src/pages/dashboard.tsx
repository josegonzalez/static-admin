import { DashboardContent } from "@/components/dashboard/dashboard-content";
import { toast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getGitHubAuthUrl, revalidateToken } from "@/lib/api";
import { isAuthenticatedToGitHub } from "@/lib/auth";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

export default function DashboardPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const [githubAuthUrl, setGithubAuthUrl] = useState<string>("");

  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        router.push("/login");
        return;
      }

      try {
        // Check if we need to revalidate the token
        if (router.query["refetch-token"] === "true") {
          const newToken = await revalidateToken();
          localStorage.setItem("token", newToken);
          router.push("/dashboard");
        }

        // Only fetch GitHub auth URL if not already authenticated
        if (!isAuthenticatedToGitHub()) {
          const url = await getGitHubAuthUrl();
          setGithubAuthUrl(url);
        }
      } catch (err) {
        toast({
          title: "Failed to authenticate with GitHub",
          description: "Error: " + err,
        });
        localStorage.removeItem("token");
        router.push("/login");
        return;
      }

      setIsLoading(false);
    };

    checkAuth();
  }, [router]);

  if (isLoading) return null;

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Dashboard</h1>
      <DashboardContent githubAuthUrl={githubAuthUrl} />
    </div>
  );
}

DashboardPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
