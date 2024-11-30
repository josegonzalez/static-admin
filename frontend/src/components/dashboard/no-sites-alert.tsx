import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { AlertCircle } from "lucide-react";
import Link from "next/link";

export function NoSitesAlert() {
  return (
    <Alert className="bg-orange-500 text-white border-orange-600">
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>No Sites Configured</AlertTitle>
      <AlertDescription>
        You have not configured any sites yet. Go to the configuration page to
        add sites.
        <p>
          <Button
            asChild
            variant="secondary"
            className="mt-2 bg-white text-orange-500 hover:bg-orange-100"
          >
            <Link href="/configuration">Go to Configuration</Link>
          </Button>
        </p>
      </AlertDescription>
    </Alert>
  );
}
