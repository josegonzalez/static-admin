"use client";

import { isAuthenticatedToGitHub } from "@/lib/auth";
import { useEffect, useState } from "react";
import { GitHubAuthAlert } from "./github-auth-alert";
import { NoSitesAlert } from "./no-sites-alert";
import { SitesList } from "./sites-list";

interface DashboardContentProps {
  githubAuthUrl: string;
}

const getConfiguredSites = () => {
  // Implement your site fetching logic here
  return Math.random() < 0.5 ? [] : [{ id: 1, name: "Example Site" }]; // Randomly return empty or non-empty array
};

export function DashboardContent({ githubAuthUrl }: DashboardContentProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [sites, setSites] = useState<Array<{ id: number; name: string }>>([]);

  useEffect(() => {
    setIsAuthenticated(isAuthenticatedToGitHub());
    setSites(getConfiguredSites());
  }, []);

  if (!isAuthenticated) {
    return <GitHubAuthAlert authUrl={githubAuthUrl} />;
  }

  if (sites.length === 0) {
    return <NoSitesAlert />;
  }

  return <SitesList sites={sites} />;
}
