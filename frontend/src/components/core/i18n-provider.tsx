"use client";

import "@/lib/i18n";

import * as React from "react";
import { useTranslation } from "react-i18next";

import { logger } from "@/lib/default-logger";

export interface I18nProviderProps {
  children: React.ReactNode;
  language?: string;
}

export function I18nProvider({ children, language = "en" }: I18nProviderProps): React.JSX.Element {
  const { i18n } = useTranslation();

  React.useEffect(() => {
    i18n.changeLanguage(language).catch(() => {
      logger.error(`Failed to change language to ${language}`);
    });
  }, [i18n, language]);

  return <React.Fragment>{children}</React.Fragment>;
}
