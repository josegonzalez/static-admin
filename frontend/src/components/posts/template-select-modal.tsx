import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { getTemplates } from "@/lib/api";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

interface Template {
  id: number;
  name: string;
}

interface TemplateSelectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  siteId: string;
  siteName: string;
}

export function TemplateSelectModal({
  open,
  onOpenChange,
  siteId,
  siteName,
}: TemplateSelectModalProps) {
  const router = useRouter();
  const [templates, setTemplates] = useState<Template[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchTemplates = async () => {
      try {
        const templates = await getTemplates();
        setTemplates(templates);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch templates",
        );
      }
    };

    if (open) {
      fetchTemplates();
    }
  }, [open]);

  const handleTemplateSelect = (templateId: number) => {
    router.push(
      `/sites/${siteId}/${siteName}/posts/new?templateId=${templateId.toString()}`,
    );
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>
            {templates.length > 0
              ? "Select a Template for the new Post"
              : "Create post without a template"}
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          {templates.length === 0 && (
            <p className="text-sm text-muted-foreground">
              Templates are a great way to create new posts with a consistent
              structure. Define templates in the{" "}
              <Link
                className="text-blue-500 hover:underline"
                href={`/sites/${siteId}/${siteName}/templates`}
              >
                site templates section.
              </Link>
            </p>
          )}
          {error ? (
            <p className="text-red-500">{error}</p>
          ) : (
            templates.map((template) => (
              <Button
                key={template.id}
                variant="outline"
                className="w-full"
                onClick={() => handleTemplateSelect(template.id)}
              >
                {template.name}
              </Button>
            ))
          )}
          {templates.length > 0 && (
            <p className="text-sm text-muted-foreground">
              Or create a new post without a template
            </p>
          )}
          <Button
            variant="outline"
            className="w-full"
            onClick={() => handleTemplateSelect(0)}
          >
            Create post without a template
          </Button>
          {templates.length > 0 && (
            <p className="text-sm text-muted-foreground">
              Templates can be defined in the{" "}
              <Link
                className="text-blue-500 hover:underline"
                href={`/sites/${siteId}/${siteName}/templates`}
              >
                site templates section.
              </Link>
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
