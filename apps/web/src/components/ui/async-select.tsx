"use client";

import * as React from "react";
import { Check, ChevronsUpDown, Loader2 } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { useDebounce } from "@/hooks/use-debounce";

interface AsyncSelectProps<T> {
  fetcher: (query: string) => Promise<T[]>;
  renderOption: (item: T) => React.ReactNode;
  getLabel: (item: T) => string;
  getValue: (item: T) => string;
  value?: string;
  onChange: (value: string, item?: T) => void;
  label: string;
  placeholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  width?: string;
  triggerClassName?: string;
  /**
   * Pre-fetched options to show initially or when query is empty
   */
  defaultOptions?: T[];
  /**
   * If true, will fetch immediately on mount/open even if query is empty
   */
  preload?: boolean;
}

export function AsyncSelect<T>({
  fetcher,
  renderOption,
  getLabel,
  getValue,
  value,
  onChange,
  label,
  placeholder,
  emptyMessage = "No results found.",
  disabled = false,
  width = "w-full",
  triggerClassName,
  defaultOptions = [],
  preload = false,
}: AsyncSelectProps<T>) {
  const [open, setOpen] = React.useState(false);
  const [query, setQuery] = React.useState("");
  const debouncedQuery = useDebounce(query, 300);
  const [fetchedOptions, setFetchedOptions] = React.useState<T[]>([]);
  const [loading, setLoading] = React.useState(false);

  const options = React.useMemo(() => {
    if (!open && defaultOptions.length > 0) {
      return defaultOptions;
    }
    return fetchedOptions;
  }, [open, defaultOptions, fetchedOptions]);

  const selectedItem = React.useMemo(() => {
    if (value && options.length > 0) {
      return options.find((item) => getValue(item) === value);
    }
    return undefined;
  }, [value, options, getValue]);

  // Fetch when debounced query changes
  React.useEffect(() => {
    const fetchOptions = async () => {
      if (!open) return;
      
      // If query is empty and we have default options, use them, unless preload is true
      if (!query && !preload && defaultOptions.length > 0) {
        setFetchedOptions(defaultOptions);
        return;
      }

      setLoading(true);
      try {
        const results = await fetcher(debouncedQuery);
        setFetchedOptions(results);
      } catch (error) {
        console.error("AsyncSelect fetch error:", error);
        setFetchedOptions([]);
      } finally {
        setLoading(false);
      }
    };

    fetchOptions();
  }, [debouncedQuery, open, fetcher, preload, query, defaultOptions]); 
  // Added query and defaultOptions to deps to handle the "empty query" check correctly
  // Note: We might want to avoid re-fetching if query is empty and we didn't require preload, 
  // but if the user cleared the input, we probably want to reset to defaultOptions or fetch all.

  const handleSelect = (currentValue: string) => {
    const item = options.find((i) => getValue(i) === currentValue);
    onChange(currentValue, item);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn("justify-between", width, triggerClassName)}
          disabled={disabled}
        >
          {selectedItem
            ? getLabel(selectedItem)
            : placeholder ?? `Select ${label}...`}
          <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className={cn("p-0", width)} align="start">
        <Command shouldFilter={false}>
          <CommandInput
            placeholder={`Search ${label}...`}
            value={query}
            onValueChange={setQuery}
          />
          <CommandList>
            {loading && (
              <div className="flex items-center justify-center py-6">
                <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
              </div>
            )}
            
            {!loading && options.length === 0 && (
              <CommandEmpty>{emptyMessage}</CommandEmpty>
            )}

            {!loading && options.length > 0 && (
              <CommandGroup>
                {options.map((item) => {
                  const itemValue = getValue(item);
                  const isSelected = value === itemValue;
                  
                  return (
                    <CommandItem
                      key={itemValue}
                      value={itemValue}
                      onSelect={handleSelect}
                    >
                      <Check
                        className={cn(
                          "mr-2 h-4 w-4",
                          isSelected ? "opacity-100" : "opacity-0"
                        )}
                      />
                      {renderOption(item)}
                    </CommandItem>
                  );
                })}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
