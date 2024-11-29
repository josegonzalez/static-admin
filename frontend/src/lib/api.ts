import { Post } from "@/types/post";
import { Site } from "@/types/site";

const API_BASE = "http://localhost:8080";

interface LoginResponse {
  token: string;
}

interface GitHubAuthURLResponse {
  url: string;
}

interface TokenResponse {
  token: string;
}

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

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem("token");
  if (token) {
    options.headers = {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    };
  }
  const response = await fetch(`${API_BASE}${url}`, options);
  if (!response.ok) throw new Error("API request failed");
  return response;
}

export async function login(email: string, password: string): Promise<string> {
  const response = await fetch(`${API_BASE}/api/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) throw new Error("Login failed");
  const data: LoginResponse = await response.json();
  return data.token;
}

export async function createAccount(
  email: string,
  password: string
): Promise<void> {
  const response = await fetch(`${API_BASE}/api/accounts`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  if (!response.ok) throw new Error("Account creation failed");
}

export async function getGitHubAuthUrl(): Promise<string> {
  const response = await fetchWithAuth("/api/github/auth-url");
  const data: GitHubAuthURLResponse = await response.json();
  return data.url;
}

export async function revalidateToken(): Promise<string> {
  const response = await fetchWithAuth("/api/auth/revalidate");
  const data: TokenResponse = await response.json();
  return data.token;
}

export async function getSites(): Promise<Site[]> {
  try {
    const response = await fetchWithAuth("/api/sites");
    if (response.status === 404) {
      throw new Error("No sites found");
    }
    const data = await response.json();
    return data;
  } catch (err) {
    if (err instanceof Error) {
      throw err;
    }
    throw new Error("Failed to fetch sites");
  }
}

export async function getGitHubOrganizations(): Promise<Organization[]> {
  const response = await fetchWithAuth("/api/github/organizations");
  return response.json();
}

export async function getGitHubRepositories(
  org: string
): Promise<Repository[]> {
  const response = await fetchWithAuth(
    `/api/github/organizations/${org}/repositories`
  );
  return response.json();
}

export async function createSite(repositoryUrl: string): Promise<void> {
  const response = await fetchWithAuth("/api/sites", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ repository_url: repositoryUrl }),
  });

  if (!response.ok) {
    throw new Error("Failed to create site");
  }
}

export async function deleteSite(id: number): Promise<void> {
  const response = await fetchWithAuth(`/api/sites/${id}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error("Failed to delete site");
  }
}

export async function getPosts(siteId: string): Promise<Post[]> {
  const response = await fetchWithAuth(`/api/sites/${siteId}/posts`);
  return response.json();
}

export async function getPost(siteId: string, postId: string): Promise<Post> {
  const response = await fetchWithAuth(`/api/sites/${siteId}/posts/${postId}`);
  return response.json();
}
