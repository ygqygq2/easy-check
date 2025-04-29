import * as React from "react";

import { config } from "@/config";
import { Metadata } from "@/types/metadata";

const metadata = {
  title: `Not found | ${config.site.name}`,
} satisfies Metadata;

export function Page(): React.JSX.Element {
  return (
    <React.Fragment>
      <meta content="width=device-width, initial-scale=1.0" name="viewport" />
      <link rel="icon" href="./src/assets/favicon.ico" />
      <title>{metadata.title}</title>
    </React.Fragment>
  );
}
