import { Box, Spinner, Text } from "@chakra-ui/react";

interface StatusViewProps {
  message: string;
  isLoading?: boolean;
}

export function StatusView({ message, isLoading = false }: StatusViewProps) {
  return (
    <Box textAlign="center" py={10} px={6}>
      {isLoading && <Spinner size="xl" />}
      <Text mt={isLoading ? 4 : 0} fontSize="lg">
        {message}
      </Text>
    </Box>
  );
}
