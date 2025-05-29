import { HStack, IconButton, Input, InputGroup } from "@chakra-ui/react";
import { Icon } from "@iconify/react";

interface SearchBarProps {
  searchTerm: string;
  setSearchTerm: (value: string) => void;
  onSearch: () => void;
}

export function SearchBar({
  searchTerm,
  setSearchTerm,
  onSearch,
}: SearchBarProps) {
  const searchIcon = (
    <IconButton colorPalette="blue">
      <Icon icon="mynaui:search" width="24" height="24" />
    </IconButton>
  );

  return (
    <HStack w="50%">
      <InputGroup endElement={searchIcon} mr={4} css={{ "& > div": { px: 0 } }}>
        <Input
          placeholder="搜索主机名称或 IP 地址"
          value={searchTerm}
          onClick={onSearch}
          onChange={(e) => setSearchTerm(e.target.value)}
        ></Input>
      </InputGroup>
    </HStack>
  );
}
