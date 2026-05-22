"use client";

import { useState, useEffect } from "react";
import { Calendar as CalendarIcon, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

interface ReminderDateTimePickerProps {
  value?: string; // ISO datetime string or empty string
  onChange: (value: string) => void;
  disabled?: boolean;
  error?: string;
  className?: string;
}

// Helper function to format date
const formatDate = (date: Date, formatStr: string): string => {
  const day = date.getDate();
  const month = date.getMonth();
  const year = date.getFullYear();
  const dayNames = [
    "Sunday",
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
  ];
  const monthNames = [
    "January",
    "February",
    "March",
    "April",
    "May",
    "June",
    "July",
    "August",
    "September",
    "October",
    "November",
    "December",
  ];

  if (formatStr === "PPP") {
    // Format: "January 1st, 2024"
    const daySuffix =
      day === 1 || day === 21 || day === 31
        ? "st"
        : day === 2 || day === 22
          ? "nd"
          : day === 3 || day === 23
            ? "rd"
            : "th";
    return `${monthNames[month]} ${day}${daySuffix}, ${year}`;
  }

  if (formatStr === "EEEE, d") {
    // Format: "Monday, 1"
    return `${dayNames[date.getDay()]}, ${day}`;
  }

  return date.toLocaleDateString();
};

export function ReminderDateTimePicker({
  value,
  onChange,
  disabled = false,
  error,
  className,
}: ReminderDateTimePickerProps) {
  const [open, setOpen] = useState(false);

  // Parse value to date and time
  const parseValue = (val?: string): { date: Date | null; time: string } => {
    if (!val) {
      return { date: null, time: "" };
    }

    try {
      const date = new Date(val);
      if (Number.isNaN(date.getTime())) {
        return { date: null, time: "" };
      }

      // Extract time in HH:mm format
      const hours = date.getHours().toString().padStart(2, "0");
      const minutes = date.getMinutes().toString().padStart(2, "0");
      const time = `${hours}:${minutes}`;

      return { date, time };
    } catch {
      return { date: null, time: "" };
    }
  };

  const { date: initialDate, time: initialTime } = parseValue(value);
  const [selectedDate, setSelectedDate] = useState<Date | null>(initialDate);
  const [selectedTime, setSelectedTime] = useState<string>(initialTime);

  // Update when value prop changes (defer updates to avoid synchronous setState in effect)
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      const parsed = parseValue(value);
      setSelectedDate(parsed.date);
      setSelectedTime(parsed.time);
    }, 0);
    return () => clearTimeout(timeoutId);
  }, [value]);

  const handleDateSelect = (newDate: Date | undefined) => {
    if (newDate) {
      // Normalize date to avoid timezone issues
      const normalizedDate = new Date(newDate);
      normalizedDate.setHours(0, 0, 0, 0);
      setSelectedDate(normalizedDate);
      updateDateTime(normalizedDate, selectedTime);
    }
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const timeValue = e.target.value;
    setSelectedTime(timeValue);
    if (selectedDate) {
      updateDateTime(selectedDate, timeValue);
    }
  };

  const updateDateTime = (date: Date, time: string) => {
    if (!date || !time) {
      onChange("");
      return;
    }

    // Parse time (HH:mm format)
    const [hours, minutes] = time.split(":").map(Number);
    if (Number.isNaN(hours) || Number.isNaN(minutes)) {
      onChange("");
      return;
    }

    // Create new date with selected date and time
    const dateTime = new Date(date);
    dateTime.setHours(hours, minutes, 0, 0);

    // Convert to ISO string
    onChange(dateTime.toISOString());
  };

  const displayText =
    selectedDate && selectedTime
      ? `${formatDate(selectedDate, "PPP")} at ${selectedTime}`
      : selectedDate
        ? formatDate(selectedDate, "PPP")
        : "Select date and time";

  return (
    <div className={cn("space-y-2", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className={cn(
              "w-full justify-start text-left font-normal h-9",
              !selectedDate && "text-muted-foreground",
              error && "border-destructive",
            )}
            disabled={disabled}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {displayText}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0" align="start">
          <div className="rounded-md border">
            <div className="flex max-sm:flex-col">
              <Calendar
                mode="single"
                selected={selectedDate || undefined}
                onSelect={handleDateSelect}
                className="p-2 sm:pe-5"
                disabled={[{ before: new Date() }]}
                showOutsideDays={false}
              />
              <div className="relative w-full max-sm:border-t sm:w-48 sm:border-s">
                <div className="p-4 space-y-3">
                  <div className="flex items-center gap-2">
                    <Clock className="h-4 w-4 text-muted-foreground" />
                    <p className="text-sm font-medium">
                      {selectedDate
                        ? formatDate(selectedDate, "EEEE, d")
                        : "Select date"}
                    </p>
                  </div>
                  <div className="space-y-2">
                    <label className="text-xs text-muted-foreground">
                      Time
                    </label>
                    <Input
                      type="time"
                      value={selectedTime}
                      onChange={handleTimeChange}
                      className="w-full"
                      placeholder="HH:mm"
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </PopoverContent>
      </Popover>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </div>
  );
}
