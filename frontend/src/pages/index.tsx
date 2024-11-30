import { Button } from "@/components/ui/button";
import Link from "next/link";
import { FaGithub, FaMarkdown } from "react-icons/fa6";
import { HiSparkles } from "react-icons/hi2";

export default function Page() {
  const features = [
    {
      name: "Simple Interface",
      description:
        "Modern, intuitive interface for managing your content without touching code.",
      href: "#",
      icon: HiSparkles,
    },
    {
      name: "GitHub Integration",
      description:
        "Seamlessly integrates with your GitHub repositories and workflow.",
      href: "#",
      icon: FaGithub,
    },
    {
      name: "Markdown Support",
      description:
        "Rich text editor with full markdown support and live preview.",
      href: "#",
      icon: FaMarkdown,
    },
  ];

  return (
    <>
      <div className="flex flex-1 flex-col gap-4 p-4">
        <header className="h-14 flex items-center">
          <Link className="flex items-center justify-center pl-8" href="/">
            <span className="font-bold">Static Admin</span>
          </Link>
          <nav className="ml-auto flex gap-4 sm:gap-6">
            <Button asChild variant="ghost" className="pr-8">
              <Link href="/login">Login</Link>
            </Button>
          </nav>
        </header>
        <div className="bg-white">
          <div className="mx-auto max-w-7xl pb-8 sm:px-6 lg:px-8">
            <div className="relative isolate overflow-hidden bg-gray-900 px-6 py-24 text-center shadow-2xl sm:rounded-3xl sm:px-16">
              <h2 className="text-balance text-4xl font-semibold tracking-tight text-white sm:text-5xl">
                Manage Your Static Sites with Ease
              </h2>
              <p className="mx-auto mt-6 max-w-xl text-pretty text-lg/8 text-gray-300">
                A modern interface for managing your static site content. Built
                for GitHub Pages and Jekyll.
              </p>
              <div className="mt-10 flex items-center justify-center gap-x-6">
                <Link
                  href="/login"
                  className="rounded-md bg-white px-3.5 py-2.5 text-sm font-semibold text-gray-900 shadow-sm hover:bg-gray-100 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-white"
                >
                  Get started
                </Link>
              </div>
              <svg
                viewBox="0 0 1024 1024"
                aria-hidden="true"
                className="absolute left-1/2 top-1/2 -z-10 size-[64rem] -translate-x-1/2 [mask-image:radial-gradient(closest-side,white,transparent)]"
              >
                <circle
                  r={512}
                  cx={512}
                  cy={512}
                  fill="url(#827591b1-ce8c-4110-b064-7cb85a0b1217)"
                  fillOpacity="0.7"
                />
                <defs>
                  <radialGradient id="827591b1-ce8c-4110-b064-7cb85a0b1217">
                    <stop stopColor="#7775D6" />
                    <stop offset={1} stopColor="#E935C1" />
                  </radialGradient>
                </defs>
              </svg>
            </div>
          </div>
        </div>

        <div className="bg-white">
          <div className="mx-auto max-w-7xl px-6 lg:px-8">
            <div className="mx-auto max-w-2xl lg:max-w-none">
              <dl className="grid max-w-xl grid-cols-1 gap-x-8 gap-y-16 lg:max-w-none lg:grid-cols-3">
                {features.map((feature) => (
                  <div key={feature.name} className="flex flex-col">
                    <dt className="flex items-center gap-x-3 text-base/7 font-semibold text-gray-900">
                      <feature.icon
                        aria-hidden="true"
                        className="size-5 flex-none text-indigo-600"
                      />
                      {feature.name}
                    </dt>
                    <dd className="mt-4 flex flex-auto flex-col text-base/7 text-gray-600">
                      <p className="flex-auto">{feature.description}</p>
                      {feature.href && feature.href !== "#" && (
                        <p className="mt-6">
                          <a
                            href={feature.href}
                            className="text-sm/6 font-semibold text-indigo-600"
                          >
                            Learn more <span aria-hidden="true">→</span>
                          </a>
                        </p>
                      )}
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
          </div>
        </div>

        <footer className="flex flex-col gap-2 sm:flex-row py-6 w-full shrink-0 items-center px-4 md:px-6 border-t">
          <p className="text-xs text-gray-500 dark:text-gray-400">
            © 2024 Static Admin. All rights reserved.
          </p>
          <nav className="sm:ml-auto flex gap-4 sm:gap-6">
            <Link
              className="text-xs hover:underline underline-offset-4"
              href="https://github.com/josegonzalez/static-admin"
            >
              GitHub
            </Link>
            <Link
              className="text-xs hover:underline underline-offset-4"
              href="https://github.com/josegonzalez/static-admin/issues"
            >
              Support
            </Link>
          </nav>
        </footer>
      </div>
    </>
  );
}
