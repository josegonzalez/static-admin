import "@/styles/globals.css";
import type { NextPage } from "next";
import { ThemeProvider } from "next-themes";
import type { AppProps } from "next/app";
import { Inter } from "next/font/google";
import { useRouter } from "next/router";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";

const inter = Inter({ subsets: ["latin"] });

type NextPageWithLayout = NextPage & {
  getLayout?: (page: ReactElement) => ReactElement;
};

type AppPropsWithLayout = AppProps & {
  Component: NextPageWithLayout;
};

export default function App({ Component, pageProps }: AppPropsWithLayout) {
  const getLayout = Component.getLayout ?? ((page) => page);
  const [ready, setReady] = useState(false);
  const router = useRouter();
  useEffect(() => {
    if (router.isReady) {
      setReady(true);
    }
  }, [router, router.isReady]);
  return (
    ready && (
      <ThemeProvider attribute="class">
        <main className={inter.className}>
          {getLayout(<Component {...pageProps} />)}
        </main>
      </ThemeProvider>
    )
  );
}
