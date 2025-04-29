import { Box, Center, Image } from "@chakra-ui/react";

import logo from "@/assets/images/logo.png";

export function Page() {
  return (
    <Box id="App" height="100%" overflow="hidden" display="flex" alignItems="center" justifyContent="center">
      <Center>
        <Image src={logo} id="logo" alt="logo" maxHeight="80vh" objectFit="contain" />
      </Center>
    </Box>
  );
}
