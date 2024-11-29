"use client";

import { getSites } from "@/lib/api";
import { isAuthenticatedToGitHub } from "@/lib/auth";
import { Site } from "@/types/site";
import { useEffect, useState } from "react";
import { GitHubAuthAlert } from "./github-auth-alert";
import { NoSitesAlert } from "./no-sites-alert";
import { RepositoriesList } from "./repositories-list";

interface DashboardContentProps {
  githubAuthUrl: string;
}

export function DashboardContent({ githubAuthUrl }: DashboardContentProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [sites, setSites] = useState<Site[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      setIsAuthenticated(isAuthenticatedToGitHub());

      try {
        const fetchedSites = await getSites();
        setSites(fetchedSites);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to fetch sites");
      }
    };

    fetchData();
  }, []);

  if (!isAuthenticated) {
    return <GitHubAuthAlert authUrl={githubAuthUrl} />;
  }

  if (error) {
    return <NoSitesAlert />;
  }

  if (sites.length === 0) {
    return <NoSitesAlert />;
  }

  return <RepositoriesList repositories={sites} />;
}
