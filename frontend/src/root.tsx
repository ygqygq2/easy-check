"use client";

import "@/styles/global.css";

import * as React from "react";

import { I18nProvider } from "@/components/core/i18n-provider";
import { Provider } from "@/components/ui/provider";

import { config } from "./config";
import type { Metadata } from "./types/metadata";

const metadata: Metadata = {
  title: `${config.site.name} - ${config.site.description}`,
};

interface RootProps {
  children: React.ReactNode;
}

export function Root({ children }: RootProps): React.JSX.Element {
  return (
    <Provider>
      <title>{metadata.title}</title>
      <I18nProvider language="en">{children}</I18nProvider>
    </Provider>
  );
}
