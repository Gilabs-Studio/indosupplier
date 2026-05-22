"use client";

import { useState } from "react";
import { Calendar as CalendarIcon, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

interface DateTimePickerProps {
  date: Date | null;
  time: string | null; // HH:mm format or null
  onDateChange: (date: Date | null, time: string | null) => void;
  disabled?: boolean;
}

// Helper function to format date (replaces date-fns format)
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

// No mock time slots - user can input freely

export function DateTimePicker({
  date,
  time,
  onDateChange,
  disabled = false,
}: DateTimePickerProps) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const [open, setOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState<Date>(date || today);
  const [selectedTime, setSelectedTime] = useState<string | null>(time);

  const handleDateSelect = (newDate: Date | undefined) => {
    if (newDate) {
      // Normalize date to avoid timezone issues
      const normalizedDate = new Date(newDate);
      normalizedDate.setHours(0, 0, 0, 0);
      setSelectedDate(normalizedDate);
      // Keep time if already selected, otherwise reset
      if (selectedTime) {
        onDateChange(normalizedDate, selectedTime);
      } else {
        onDateChange(normalizedDate, null);
      }
    }
  };

  const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const timeValue = e.target.value;
    setSelectedTime(timeValue);
    if (selectedDate) {
      onDateChange(selectedDate, timeValue || null);
    }
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          size="sm"
          className="w-full justify-start text-left font-normal"
          disabled={disabled}
        >
          <CalendarIcon className="mr-2 h-4 w-4" />
          {date && time
            ? `${formatDate(date, "PPP")} at ${time}`
            : date
              ? formatDate(date, "PPP")
              : "Select date and time"}
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <div className="rounded-md border">
          <div className="flex max-sm:flex-col">
            <Calendar
              mode="single"
              selected={selectedDate}
              onSelect={handleDateSelect}
              className="p-2 sm:pe-5"
              disabled={[{ before: today }]}
              showOutsideDays={false}
            />
            <div className="relative w-full max-sm:border-t sm:w-48 sm:border-s">
              <div className="p-4 space-y-3">
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-muted-foreground" />
                  <p className="text-sm font-medium">
                    {formatDate(selectedDate, "EEEE, d")}
                  </p>
                </div>
                <div className="space-y-2">
                  <label className="text-xs text-muted-foreground">Time</label>
                  <Input
                    type="time"
                    value={selectedTime || ""}
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
  );
}
