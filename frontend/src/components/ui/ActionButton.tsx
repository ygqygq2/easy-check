import { Button } from "@chakra-ui/react";

import { useColorModeValue } from "./color-mode";

interface ActionButtonProps {
  label: string;
  onClick: () => void;
}

const ActionButton = ({ label, onClick }: ActionButtonProps) => {
  const buttonBg = useColorModeValue("gray.200", "gray.700");
  const buttonColor = useColorModeValue("gray.800", "white");

  return (
    <Button
      bg={buttonBg}
      color={buttonColor}
      _hover={{ bg: useColorModeValue("gray.300", "gray.600") }}
      onClick={onClick}
    >
      {label}
    </Button>
  );
};

export default ActionButton;
