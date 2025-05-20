"use client";
import { Button } from "@chakra-ui/react";
import { Flex, Image } from "@chakra-ui/react";

import logo from "@/assets/images/logo.png";

export function Page() {
  return (
    <>
      <Flex
        id="App"
        height="90vh"
        width="100%"
        overflow="hidden"
        direction="column"
        alignItems="center"
        justifyContent="center"
      >
        <Image src={logo} id="logo" alt="logo" maxHeight="80vh" maxWidth="90%" objectFit="contain" />
      </Flex>
    </>
  );
}
