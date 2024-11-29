import { RepositoriesList } from "@/components/dashboard/repositories-list";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import DashboardLayout from "@/layouts/dashboard-layout";
import { deleteSite, getSites } from "@/lib/api";
import { AlertCircle } from "lucide-react";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";
interface Site {
  id: number;
  default_branch: string;
  description: string;
  name: string;
  private: boolean;
  url: string;
}

export default function ConfigurationPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const [sites, setSites] = useState<Site[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        router.push("/login");
        return;
      }

      try {
        const fetchedSites = await getSites();
        setSites(fetchedSites);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("An unexpected error occurred while fetching sites");
        }
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [router]);
  const handleDelete = async (id: number) => {
    // Show confirmation dialog
    if (
      !window.confirm(
        "Are you sure you want to delete this site? You will need to re-add the site to update posts in the future."
      )
    ) {
      return;
    }

    try {
      await deleteSite(id);
      const updatedSites = await getSites();
      setSites(updatedSites);
      setError(null);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to delete site");
      }
    }
  };

  if (isLoading) return null;

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Configuration</h1>
      {error ? (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : sites.length === 0 ? (
        <Alert className="bg-blue-500 text-white border-blue-600">
          <AlertCircle className="h-4 w-4 text-blue-600 bg-blue-500" />
          <AlertTitle>No sites configured</AlertTitle>
          <AlertDescription>
            Sites must be configured before you can use edit posts.
            <p>
              <Button
                onClick={() => router.push("/configuration/sites")}
                variant="secondary"
                className="mt-2 bg-white text-blue-500 hover:bg-blue-100"
              >
                Add a site
              </Button>
            </p>
          </AlertDescription>
        </Alert>
      ) : (
        <>
          <RepositoriesList repositories={sites} onDelete={handleDelete} />
          <p>
            <Button
              onClick={() => router.push("/configuration/sites")}
              variant="secondary"
              className="mt-2 bg-white text-blue-500 hover:bg-blue-100"
            >
              Add a site
            </Button>
          </p>
        </>
      )}
    </div>
  );
}

ConfigurationPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
