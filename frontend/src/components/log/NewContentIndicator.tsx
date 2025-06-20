import { Box, Text } from "@chakra-ui/react";

interface NewContentIndicatorProps {
  onScrollToBottom: () => void;
}

export const NewContentIndicator = ({
  onScrollToBottom,
}: NewContentIndicatorProps) => {
  return (
    <Box
      position="absolute"
      bottom="20px"
      right="20px"
      bg="blue.500"
      color="white"
      p={2}
      borderRadius="md"
      cursor="pointer"
      onClick={onScrollToBottom}
      boxShadow="md"
      _hover={{ bg: "blue.600" }}
    >
      <Text fontSize="sm">新内容 ↓</Text>
    </Box>
  );
};
