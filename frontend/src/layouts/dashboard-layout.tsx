import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { toast } from "@/hooks/use-toast";
import { getSites } from "@/lib/api";
import { Site } from "@/types/site";
import { LogOut } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [sites, setSites] = useState<Site[]>([]);

  useEffect(() => {
    const fetchSites = async () => {
      try {
        const fetchedSites = await getSites();
        setSites(fetchedSites);
      } catch (err) {
        toast({
          title: "Failed to fetch sites",
          description: err instanceof Error ? err.message : "Unknown error",
        });
      }
    };

    fetchSites();
  }, []);

  return (
    <SidebarProvider>
      <div className="flex h-screen">
        <Sidebar>
          <SidebarHeader>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton>
                  <Link href="/dashboard">Static Admin</Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarHeader>
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupContent>
                <SidebarMenu>
                  <SidebarMenuItem>
                    <SidebarMenuButton asChild>
                      <a>Posts</a>
                    </SidebarMenuButton>
                    {sites.map((site) => (
                      <SidebarMenuSub key={site.id}>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton asChild>
                            <Link href={`/sites/${site.id}/${site.name}/posts`}>
                              {site.name}
                            </Link>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      </SidebarMenuSub>
                    ))}
                    {sites.length === 0 && (
                      <SidebarMenuSub>
                        <SidebarMenuSubItem>
                          <SidebarMenuSubButton asChild>
                            <Link href="/configuration/sites">
                              Add new site...
                            </Link>
                          </SidebarMenuSubButton>
                        </SidebarMenuSubItem>
                      </SidebarMenuSub>
                    )}
                  </SidebarMenuItem>
                  <SidebarMenuItem>
                    <SidebarMenuButton asChild>
                      <Link href="/configuration">Configuration</Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarFooter>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <Link href="/logout" className="flex items-center gap-2">
                    <span>Logout</span>
                    <LogOut className="h-4 w-4" />
                  </Link>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarFooter>
        </Sidebar>
        <main className="flex-1 p-6 overflow-auto">{children}</main>
      </div>
    </SidebarProvider>
  );
}
