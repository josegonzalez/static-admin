import {
  DEFAULT_FIELDS,
  TemplateForm,
} from "@/components/templates/template-form";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { getTemplate, updateTemplate } from "@/lib/api";
import { useRouter } from "next/router";
import { useEffect, useState } from "react";

export default function EditTemplatePage() {
  const router = useRouter();
  const { toast } = useToast();
  const { siteId, siteName, templateId } = router.query;
  const [templateName, setTemplateName] = useState("");
  const [fields, setFields] = useState(DEFAULT_FIELDS);
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isAddingField, setIsAddingField] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const fetchTemplate = async () => {
      try {
        if (typeof templateId !== "string") return;
        const template = await getTemplate(templateId);
        setTemplateName(template.name);
        setFields(template.fields);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to fetch template",
        );
      } finally {
        setIsLoading(false);
      }
    };

    if (templateId) {
      fetchTemplate();
    }
  }, [templateId]);

  const handleSave = async () => {
    if (!templateName.trim()) {
      setError("Template name is required");
      return;
    }

    if (typeof templateId !== "string") {
      setError("Template ID is required");
      return;
    }

    setIsSaving(true);
    try {
      await updateTemplate(templateId, {
        name: templateName,
        fields,
      });

      toast({
        title: "Success",
        description: "Template updated successfully",
      });

      router.push(`/sites/${siteId}/${siteName}/templates`);
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to update template",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  if (isLoading) return null;

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Edit Template</h1>
      </div>

      <TemplateForm
        templateName={templateName}
        setTemplateName={setTemplateName}
        fields={fields}
        setFields={setFields}
        error={error}
        setError={setError}
        isAddingField={isAddingField}
        setIsAddingField={setIsAddingField}
        isSaving={isSaving}
        onSave={handleSave}
        onCancel={() => router.push(`/sites/${siteId}/${siteName}/templates`)}
        saveButtonText="Save Changes"
      />
    </div>
  );
}

EditTemplatePage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
