import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Site } from "@/types/site";
import { Lock, Unlock } from "lucide-react";
import Link from "next/link";
import { Button } from "../ui/button";

interface RepositoriesListProps {
  repositories: Site[];
  onDelete?: (id: number) => void;
}

export function RepositoriesList({
  repositories,
  onDelete,
}: RepositoriesListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {repositories.map((repository) => (
        <Card key={repository.id} className="w-[380px]">
          <CardHeader>
            <CardTitle>Name: {repository.name}</CardTitle>
            <CardDescription>Site ID: {repository.id}</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <dl className="grid grid-cols-[auto,1fr] gap-x-4 gap-y-2">
              <dt className="text-sm font-medium text-muted-foreground">
                Repository
              </dt>
              <dd className="text-sm">
                <Link
                  href={repository.url}
                  className="text-blue-500 hover:underline"
                >
                  GitHub
                </Link>
              </dd>

              <dt className="text-sm font-medium text-muted-foreground">
                Description
              </dt>
              <dd className="text-sm">{repository.description}</dd>

              <dt className="text-sm font-medium text-muted-foreground">
                Visibility
              </dt>
              <dd className="text-sm flex items-center gap-1">
                {repository.private ? (
                  <>
                    <Lock className="h-4 w-4" /> Private
                  </>
                ) : (
                  <>
                    <Unlock className="h-4 w-4" /> Public
                  </>
                )}
              </dd>

              <dt className="text-sm font-medium text-muted-foreground">
                Branch
              </dt>
              <dd className="text-sm">{repository.default_branch}</dd>

              <dt className="text-sm font-medium text-muted-foreground">
                Posts
              </dt>
              <dd className="text-sm">
                <Link
                  href={`/sites/${repository.id}/${repository.name}/posts`}
                  className="text-blue-500 hover:underline"
                >
                  View
                </Link>
              </dd>

              <dt className="text-sm font-medium text-muted-foreground">
                Templates
              </dt>
              <dd className="text-sm">
                <Link
                  href={`/sites/${repository.id}/${repository.name}/templates`}
                  className="text-blue-500 hover:underline"
                >
                  View
                </Link>
              </dd>
            </dl>

            {onDelete && (
              <div className="pt-4">
                <Button
                  onClick={() => onDelete(repository.id)}
                  className="bg-red-500 hover:text-red-700"
                  aria-label={`Delete ${repository.name}`}
                >
                  Delete
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
