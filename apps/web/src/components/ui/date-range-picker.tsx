"use client";

import { useState } from "react";
import { Calendar as CalendarIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import type { DateRange } from "react-day-picker";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

interface DateRangePickerProps {
  readonly dateRange: DateRange | undefined;
  readonly onDateChange: (dateRange: DateRange | undefined) => void;
  readonly disabled?: boolean;
  readonly disabledDays?: (date: Date) => boolean;
}

// Helper functions to replace date-fns
const subDays = (date: Date, days: number) => {
  const result = new Date(date);
  result.setDate(result.getDate() - days);
  return result;
};

const subMonths = (date: Date, months: number) => {
  const result = new Date(date);
  result.setMonth(result.getMonth() - months);
  return result;
};

const subYears = (date: Date, years: number) => {
  const result = new Date(date);
  result.setFullYear(result.getFullYear() - years);
  return result;
};

const startOfMonth = (date: Date) => {
  const result = new Date(date);
  result.setDate(1);
  result.setHours(0, 0, 0, 0);
  return result;
};

const endOfMonth = (date: Date) => {
  const result = new Date(date);
  result.setMonth(result.getMonth() + 1);
  result.setDate(0);
  result.setHours(23, 59, 59, 999);
  return result;
};

const startOfYear = (date: Date) => {
  const result = new Date(date);
  result.setMonth(0);
  result.setDate(1);
  result.setHours(0, 0, 0, 0);
  return result;
};

const endOfYear = (date: Date) => {
  const result = new Date(date);
  result.setMonth(11);
  result.setDate(31);
  result.setHours(23, 59, 59, 999);
  return result;
};

export function DateRangePicker({
  dateRange,
  onDateChange,
  disabled = false,
  disabledDays,
}: DateRangePickerProps) {
  const today = new Date();
  today.setHours(23, 59, 59, 999);

  const yesterday = {
    from: subDays(today, 1),
    to: subDays(today, 1),
  };
  const last7Days = {
    from: subDays(today, 6),
    to: today,
  };
  const last30Days = {
    from: subDays(today, 29),
    to: today,
  };
  const monthToDate = {
    from: startOfMonth(today),
    to: today,
  };
  const lastMonth = {
    from: startOfMonth(subMonths(today, 1)),
    to: endOfMonth(subMonths(today, 1)),
  };
  const yearToDate = {
    from: startOfYear(today),
    to: today,
  };
  const lastYear = {
    from: startOfYear(subYears(today, 1)),
    to: endOfYear(subYears(today, 1)),
  };

  const [month, setMonth] = useState(today);
  const [open, setOpen] = useState(false);

  const handlePresetClick = (range: DateRange) => {
    onDateChange(range);
    setMonth(range.to || range.from || today);
  };

  // Check if a date range contains any disabled dates
  const isRangeDisabled = (range: DateRange): boolean => {
    if (!disabledDays) return false;
    if (!range.from) return false;

    const start = new Date(range.from);
    const end = range.to ? new Date(range.to) : start;

    // Check each day in the range
    const current = new Date(start);
    while (current <= end) {
      if (disabledDays(new Date(current))) {
        return true;
      }
      current.setDate(current.getDate() + 1);
    }
    return false;
  };

  const formatDateRange = (range: DateRange | undefined): string => {
    if (!range?.from) return "Select date range";
    if (range.to) {
      return `${range.from.toLocaleDateString("id-ID")} - ${range.to.toLocaleDateString("id-ID")}`;
    }
    return `${range.from.toLocaleDateString("id-ID")} - ...`;
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="h-9 w-[200px] justify-start text-left font-normal"
          disabled={disabled}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {formatDateRange(dateRange)}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="rounded-md border">
          <div className="flex max-sm:flex-col">
            <div className="relative py-4 max-sm:order-1 max-sm:border-t sm:w-32">
              <div className="h-full sm:border-e">
                <div className="flex flex-col px-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled({ from: today, to: today })}
                    onClick={() => {
                      handlePresetClick({
                        from: today,
                        to: today,
                      });
                    }}
                  >
                    Today
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(yesterday)}
                    onClick={() => handlePresetClick(yesterday)}
                  >
                    Yesterday
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(last7Days)}
                    onClick={() => handlePresetClick(last7Days)}
                  >
                    Last 7 days
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(last30Days)}
                    onClick={() => handlePresetClick(last30Days)}
                  >
                    Last 30 days
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(monthToDate)}
                    onClick={() => handlePresetClick(monthToDate)}
                  >
                    Month to date
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(lastMonth)}
                    onClick={() => handlePresetClick(lastMonth)}
                  >
                    Last month
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(yearToDate)}
                    onClick={() => handlePresetClick(yearToDate)}
                  >
                    Year to date
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="w-full justify-start"
                    disabled={isRangeDisabled(lastYear)}
                    onClick={() => handlePresetClick(lastYear)}
                  >
                    Last year
                  </Button>
                </div>
              </div>
            </div>
            <Calendar
              mode="range"
              selected={dateRange}
              onSelect={(newDate: unknown) => {
                // Type assertion: when mode="range", onSelect receives DateRange | undefined
                const range = newDate as DateRange | undefined;
                if (range) {
                  // Normalize dates to avoid timezone issues
                  const normalizedRange: DateRange = {
                    from: range.from
                      ? (() => {
                          const d = new Date(range.from);
                          d.setHours(0, 0, 0, 0);
                          return d;
                        })()
                      : undefined,
                    to: range.to
                      ? (() => {
                          const d = new Date(range.to);
                          d.setHours(0, 0, 0, 0);
                          return d;
                        })()
                      : undefined,
                  };
                  onDateChange(normalizedRange);
                }
              }}
              month={month}
              onMonthChange={(newMonth) => {
                if (newMonth) {
                  setMonth(newMonth);
                }
              }}
              disabled={disabledDays}
              className="p-2"
            />
          </div>
        </div>
      </PopoverContent>
    </Popover>
  );
}
