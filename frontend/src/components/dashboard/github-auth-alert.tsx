import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { AlertCircle } from "lucide-react";

interface GitHubAuthAlertProps {
  authUrl: string;
}

export function GitHubAuthAlert({ authUrl }: GitHubAuthAlertProps) {
  const handleAuth = () => {
    window.location.href = authUrl;
  };

  return (
    <Alert
      variant="destructive"
      className="bg-white-500 text-white border-green-600"
    >
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>Authentication Required</AlertTitle>
      <AlertDescription>
        You are not authenticated to GitHub. Please authenticate to access the
        dashboard.
        <p>
          <Button
            onClick={handleAuth}
            variant="secondary"
            className="mt-2 bg-white text-green-500 hover:bg-green-100"
          >
            Authenticate with GitHub
          </Button>
        </p>
      </AlertDescription>
    </Alert>
  );
}
