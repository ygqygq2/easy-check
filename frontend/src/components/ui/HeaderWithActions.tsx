import { Flex, Text } from "@chakra-ui/react";
import React from "react";

interface HeaderWithActionsProps {
  title: string;
  actions: React.ReactNode;
}

export function HeaderWithActions({ title, actions }: HeaderWithActionsProps) {
  return (
    <Flex justify="space-between" align="center" mb={4}>
      <Text fontSize="xl">{title}</Text>
      <Flex gap={4}>{actions}</Flex>
    </Flex>
  );
}
