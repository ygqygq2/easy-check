import type { RouteObject } from "react-router-dom";
import { Outlet } from "react-router-dom";

import { Layout as HomeLayout } from "@/components/layout/Layout";
import { Page as HomePage } from "@/pages/home";
import { Page as NotFoundPage } from "@/pages/not-found";

export const routes: RouteObject[] = [
  {
    element: (
      <HomeLayout>
        <Outlet />
      </HomeLayout>
    ),
    children: [{ index: true, element: <HomePage /> }],
  },
  { path: "*", element: <NotFoundPage /> },
];
