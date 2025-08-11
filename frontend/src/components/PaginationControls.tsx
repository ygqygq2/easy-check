import { ButtonGroup, Flex, IconButton } from "@chakra-ui/react";
import { Pagination } from "@chakra-ui/react";
import { Icon } from "@iconify/react";

interface PaginationControlsProps {
  page: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

export function PaginationControls({
  page,
  total,
  pageSize,
  onPageChange,
}: PaginationControlsProps) {
  const totalPages = Math.ceil(total / pageSize);
  console.log(
    `PaginationControls: page=${page}, total=${total}, pageSize=${pageSize}, totalPages=${totalPages}`
  );

  return (
    <Pagination.Root
      count={total}
      pageSize={pageSize}
      page={page}
      onPageChange={(e) => onPageChange(e.page)}
    >
      <Flex justify="flex-end" mt="4">
        <ButtonGroup variant="ghost" size="sm">
          <Pagination.PrevTrigger asChild>
            <IconButton>
              <Icon icon="line-md:chevron-small-left" width="24" height="24" />
            </IconButton>
          </Pagination.PrevTrigger>

          <Pagination.Items
            render={(page) => (
              <IconButton variant={{ base: "ghost", _selected: "outline" }}>
                {page.value}
              </IconButton>
            )}
          />

          <Pagination.NextTrigger asChild>
            <IconButton>
              <Icon icon="line-md:chevron-small-right" width="24" height="24" />
            </IconButton>
          </Pagination.NextTrigger>
        </ButtonGroup>
      </Flex>
    </Pagination.Root>
  );
}
