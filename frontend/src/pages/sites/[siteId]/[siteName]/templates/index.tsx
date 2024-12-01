import { Alert, AlertDescription } from "@/components/ui/alert";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { deleteTemplate, getTemplates } from "@/lib/api";
import { Template } from "@/types/template";
import Link from "next/link";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

export default function TemplatesPage() {
  const router = useRouter();
  const { toast } = useToast();
  const { siteId, siteName } = router.query;
  const [templates, setTemplates] = useState<Template[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedTemplates, setSelectedTemplates] = useState<Set<number>>(
    new Set(),
  );
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  useEffect(() => {
    const fetchTemplates = async () => {
      try {
        const fetchedTemplates = await getTemplates();
        setTemplates(fetchedTemplates);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch templates",
        );
      } finally {
        setIsLoading(false);
      }
    };

    if (siteId && siteName) {
      fetchTemplates();
    }
  }, [siteId, siteName]);

  const filteredTemplates = templates.filter((template) =>
    template.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedTemplates(new Set(filteredTemplates.map((t) => t.id)));
    } else {
      setSelectedTemplates(new Set());
    }
  };

  const handleSelectTemplate = (templateId: number, checked: boolean) => {
    const newSelected = new Set(selectedTemplates);
    if (checked) {
      newSelected.add(templateId);
    } else {
      newSelected.delete(templateId);
    }
    setSelectedTemplates(newSelected);
  };

  const handleDelete = async () => {
    try {
      await Promise.all(
        Array.from(selectedTemplates).map((id) => deleteTemplate(id)),
      );
      setTemplates(templates.filter((t) => !selectedTemplates.has(t.id)));
      setSelectedTemplates(new Set());
      toast({
        title: "Success",
        description: "Templates deleted successfully",
      });
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to delete templates",
        variant: "destructive",
      });
    } finally {
      setShowDeleteDialog(false);
    }
  };

  const handleDeleteSingle = async (templateId: number) => {
    try {
      await deleteTemplate(templateId);
      setTemplates(templates.filter((t) => t.id !== templateId));
      toast({
        title: "Success",
        description: "Template deleted successfully",
      });
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to delete template",
        variant: "destructive",
      });
    }
  };

  if (isLoading) return null;

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Templates</h1>
      </div>

      <div className="flex justify-between items-center gap-4">
        <Input
          placeholder="Filter templates..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-sm"
        />
        {selectedTemplates.size > 0 && (
          <Button
            variant="destructive"
            onClick={() => setShowDeleteDialog(true)}
          >
            Delete Selected ({selectedTemplates.size})
          </Button>
        )}
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">
                <Checkbox
                  checked={
                    filteredTemplates.length > 0 &&
                    filteredTemplates.every((t) => selectedTemplates.has(t.id))
                  }
                  onCheckedChange={handleSelectAll}
                />
              </TableHead>
              <TableHead>ID</TableHead>
              <TableHead>Name</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredTemplates.map((template) => (
              <TableRow key={template.id}>
                <TableCell>
                  <Checkbox
                    checked={selectedTemplates.has(template.id)}
                    onCheckedChange={(checked) =>
                      handleSelectTemplate(template.id, checked as boolean)
                    }
                  />
                </TableCell>
                <TableCell>{template.id}</TableCell>
                <TableCell>{template.name}</TableCell>
                <TableCell>
                  <div className="flex gap-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-blue-500 hover:text-blue-700"
                      asChild
                    >
                      <Link
                        href={`/sites/${siteId}/${siteName}/templates/edit/${template.id}`}
                      >
                        Edit
                      </Link>
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="text-red-500 hover:text-red-700"
                      onClick={() => {
                        if (
                          window.confirm(
                            "Are you sure you want to delete this template?",
                          )
                        ) {
                          handleDeleteSingle(template.id);
                        }
                      }}
                    >
                      Delete
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex">
        <Button asChild>
          <Link href={`/sites/${siteId}/${siteName}/templates/new`}>
            Create Template
          </Link>
        </Button>
      </div>

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
            <AlertDialogDescription>
              This will permanently delete {selectedTemplates.size} template
              {selectedTemplates.size === 1 ? "" : "s"}. This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete}>Delete</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

TemplatesPage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
