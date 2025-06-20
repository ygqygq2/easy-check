import type { RouteObject } from "react-router-dom";
import { Outlet } from "react-router-dom";

import { Layout as HomeLayout } from "@/components/layout/Layout";

export const route: RouteObject = {
  path: "/",
  element: (
    <HomeLayout>
      <Outlet />
    </HomeLayout>
  ),
  children: [
    {
      index: true,
      lazy: async () => {
        const { Page } = await import("@/pages/home");
        return { Component: Page };
      },
    },
  ],
};
