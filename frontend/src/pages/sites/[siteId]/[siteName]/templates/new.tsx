import {
  DEFAULT_FIELDS,
  TemplateForm,
} from "@/components/templates/template-form";
import { useToast } from "@/hooks/use-toast";
import DashboardLayout from "@/layouts/dashboard-layout";
import { createTemplate } from "@/lib/api";
import { useRouter } from "next/router";
import { useState } from "react";

export default function NewTemplatePage() {
  const router = useRouter();
  const { toast } = useToast();
  const { siteId, siteName } = router.query;
  const [templateName, setTemplateName] = useState("");
  const [fields, setFields] = useState(DEFAULT_FIELDS);
  const [error, setError] = useState<string | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isAddingField, setIsAddingField] = useState(false);

  const handleSave = async () => {
    if (!templateName.trim()) {
      setError("Template name is required");
      return;
    }

    setIsSaving(true);
    try {
      await createTemplate({
        name: templateName,
        fields,
      });

      toast({
        title: "Success",
        description: "Template created successfully",
      });

      router.push(`/sites/${siteId}/${siteName}/templates`);
    } catch (err) {
      toast({
        title: "Error",
        description:
          err instanceof Error ? err.message : "Failed to create template",
        variant: "destructive",
      });
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold">Create New Template</h1>
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
        saveButtonText="Create Template"
      />
    </div>
  );
}

NewTemplatePage.getLayout = function getLayout(page: React.ReactNode) {
  return <DashboardLayout>{page}</DashboardLayout>;
};
