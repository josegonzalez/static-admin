import { RepositoriesList } from "@/components/dashboard/repositories-list";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
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
import {
  createSite,
  deleteSite,
  getGitHubOrganizations,
  getGitHubRepositories,
  getSites,
} from "@/lib/api";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

interface Organization {
  login: string;
  name: string;
  url: string;
}

interface Repository {
  name: string;
  full_name: string;
  description: string;
  url: string;
  html_url: string;
  private: boolean;
  owner: string;
  default_branch: string;
}

interface Site {
  id: number;
  default_branch: string;
  description: string;
  name: string;
  private: boolean;
  url: string;
}

export default function SitesPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(true);
  const [sites, setSites] = useState<Site[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [selectedOrg, setSelectedOrg] = useState<string>("");
  const [selectedRepos, setSelectedRepos] = useState<Set<string>>(new Set());
  const [isSaving, setIsSaving] = useState(false);
  const [searchFilter, setSearchFilter] = useState("");

  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem("token");
      if (!token) {
        router.push("/login");
        return;
      }

      try {
        const [fetchedSites, fetchedOrgs] = await Promise.all([
          getSites(),
          getGitHubOrganizations(),
        ]);
        setSites(fetchedSites);
        setOrganizations(fetchedOrgs);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("An unexpected error occurred while fetching data");
        }
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, [router]);

  useEffect(() => {
    const fetchRepositories = async () => {
      if (!selectedOrg) {
        setRepositories([]);
        setSelectedRepos(new Set());
        return;
      }

      try {
        const repos = await getGitHubRepositories(selectedOrg);
        setRepositories(repos);
        setSelectedRepos(new Set());
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("Failed to fetch repositories");
        }
      }
    };

    fetchRepositories();
  }, [selectedOrg]);

  const handleSave = async () => {
    if (selectedRepos.size === 0) return;

    setIsSaving(true);
    try {
      await Promise.all(
        Array.from(selectedRepos).map((repoUrl) => createSite(repoUrl))
      );
      const updatedSites = await getSites();
      setSites(updatedSites);
      setSelectedOrg("");
      setSelectedRepos(new Set());
      setError(null);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to save sites");
      }
    } finally {
      setIsSaving(false);
    }
  };

  const toggleAllRepositories = (checked: boolean) => {
    if (checked) {
      setSelectedRepos(new Set(repositories.map((repo) => repo.html_url)));
    } else {
      setSelectedRepos(new Set());
    }
  };

  const toggleRepository = (url: string) => {
    const newSelected = new Set(selectedRepos);
    if (newSelected.has(url)) {
      newSelected.delete(url);
    } else {
      newSelected.add(url);
    }
    setSelectedRepos(newSelected);
  };

  const filteredRepositories = repositories.filter((repo) =>
    repo.name.toLowerCase().includes(searchFilter.toLowerCase())
  );

  const handleDelete = async (id: number) => {
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
    <div className="space-y-8">
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Configured Sites</h1>
        {error ? (
          <Alert variant="destructive">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        ) : sites.length === 0 ? (
          <Alert>
            <AlertDescription>No sites configured yet.</AlertDescription>
          </Alert>
        ) : (
          <RepositoriesList repositories={sites} onDelete={handleDelete} />
        )}
      </div>

      <div className="space-y-6">
        <h2 className="text-2xl font-bold">Add New Sites</h2>
        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium">Organization</label>
            <Select value={selectedOrg} onValueChange={setSelectedOrg}>
              <SelectTrigger>
                <SelectValue placeholder="Select an organization" />
              </SelectTrigger>
              <SelectContent>
                {organizations.map((org) => (
                  <SelectItem key={org.login} value={org.login}>
                    {org.name || org.login}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {selectedOrg && repositories.length > 0 && (
            <div className="space-y-4">
              <div>
                <label htmlFor="search" className="text-sm font-medium">
                  Filter Repositories
                </label>
                <Input
                  id="search"
                  placeholder="Type to filter..."
                  value={searchFilter}
                  onChange={(e) => setSearchFilter(e.target.value)}
                  className="max-w-sm"
                />
              </div>
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-12">
                        <Checkbox
                          checked={selectedRepos.size === repositories.length}
                          onCheckedChange={toggleAllRepositories}
                        />
                      </TableHead>
                      <TableHead>Repository</TableHead>
                      <TableHead>Description</TableHead>
                      <TableHead>Visibility</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredRepositories.map((repo) => (
                      <TableRow key={repo.full_name}>
                        <TableCell>
                          <Checkbox
                            checked={selectedRepos.has(repo.html_url)}
                            onCheckedChange={() =>
                              toggleRepository(repo.html_url)
                            }
                          />
                        </TableCell>
                        <TableCell>{repo.name}</TableCell>
                        <TableCell>{repo.description}</TableCell>
                        <TableCell>
                          {repo.private ? "Private" : "Public"}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            </div>
          )}

          <Button
            onClick={handleSave}
            disabled={selectedRepos.size === 0 || isSaving}
          >
            {isSaving
              ? "Saving..."
              : `Save ${selectedRepos.size} Selected ${
                  selectedRepos.size === 1 ? "Site" : "Sites"
                }`}
          </Button>
        </div>
      </div>
    </div>
  );
}

SitesPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
