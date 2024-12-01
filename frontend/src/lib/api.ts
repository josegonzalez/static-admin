import { FrontmatterField } from "@/types/frontmatter";
import { Post } from "@/types/post";
import { SavePostResponse } from "@/types/save-post-response";
import { Site } from "@/types/site";
import { Template } from "@/types/template";

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

function host(): string {
  const value = process.env.NEXT_PUBLIC_API_HOSTNAME;
  if (value !== undefined && value !== "") {
    return value;
  }

  if (typeof window !== "undefined") {
    if (window.location.host === "localhost:3000") {
      return "localhost:8080";
    }
  }
  return window.location.host;
}

function scheme(): string {
  const value = process.env.NEXT_PUBLIC_API_SCHEME;
  if (value !== undefined && value !== "") {
    return value;
  }

  if (typeof window === "undefined") {
    return "https";
  }

  if (window.location.protocol === "http:") {
    return "http";
  }

  return "https";
}

function baseUrl(): string {
  return `${scheme()}://${host()}`;
}

async function fetchWithAuth(url: string, options: RequestInit = {}) {
  const token = localStorage.getItem("token");
  if (token) {
    options.headers = {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    };
  }
  const response = await fetch(`${baseUrl()}${url}`, options);
  if (!response.ok) throw new Error("API request failed");
  return response;
}

export async function login(email: string, password: string): Promise<string> {
  const response = await fetch(`${baseUrl()}/api/login`, {
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
  password: string,
): Promise<void> {
  const response = await fetch(`${baseUrl()}/api/accounts`, {
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
  org: string,
): Promise<Repository[]> {
  const response = await fetchWithAuth(
    `/api/github/organizations/${org}/repositories`,
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

export async function savePost(
  siteId: string,
  postId: string,
  post: Post,
): Promise<SavePostResponse> {
  let url = `/api/sites/${siteId}/posts/${postId}`;
  let method = "POST";
  if (postId === "new") {
    url = `/api/sites/${siteId}/posts`;
    method = "PUT";
  }

  const response = await fetchWithAuth(url, {
    method: method,
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(post),
  });

  if (!response.ok) {
    throw new Error("Failed to save post");
  }
  return response.json();
}

export async function getTemplates(): Promise<Template[]> {
  const response = await fetchWithAuth("/api/templates");
  return response.json();
}

export async function deleteTemplate(id: number): Promise<void> {
  const response = await fetchWithAuth(`/api/templates/${id}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error("Failed to delete template");
  }
}

export async function createTemplate(template: {
  name: string;
  fields: FrontmatterField[];
}): Promise<void> {
  const response = await fetchWithAuth("/api/templates", {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(template),
  });

  if (!response.ok) {
    throw new Error("Failed to create template");
  }
}

export async function getTemplate(id: string): Promise<{
  name: string;
  fields: FrontmatterField[];
}> {
  const response = await fetchWithAuth(`/api/templates/${id}`);
  return response.json();
}

export async function updateTemplate(
  id: string,
  template: {
    name: string;
    fields: FrontmatterField[];
  },
): Promise<void> {
  const response = await fetchWithAuth(`/api/templates/${id}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(template),
  });

  if (!response.ok) {
    throw new Error("Failed to update template");
  }
}
