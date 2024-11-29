import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import Link from "next/link";
import { Button } from "../ui/button";

interface Repository {
  id: number;
  name: string;
  url: string;
}

interface RepositoriesListProps {
  repositories: Repository[];
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
            <div className="mb-2 grid items-start last:mb-0 last:pb-0">
              <div className="space-y-1">
                <p className="text-sm font-medium leading-none">
                  Repository Link
                </p>
                <p className="text-sm text-muted-foreground">
                  <Link href={repository.url}>Github</Link>
                </p>
              </div>
            </div>
            {onDelete && (
              <div className="mb-2 grid items-start last:mb-0 last:pb-0">
                <div className="space-y-1">
                  <p className="text-sm font-medium leading-none">
                    Delete this?
                  </p>
                  <p className="text-sm">
                    <Button
                      onClick={() => onDelete(repository.id)}
                      className="bg-red-500 hover:text-red-700"
                      aria-label={`Delete ${repository.name}`}
                    >
                      Delete
                    </Button>
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
