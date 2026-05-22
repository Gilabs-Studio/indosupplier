import * as React from "react";
import { ChevronLeftIcon, ChevronRightIcon } from "lucide-react";
import { DayPicker, type DayPickerProps } from "react-day-picker";

import { cn, parseLocalDate } from "@/lib/utils";
import { buttonVariants } from "@/components/ui/button";

type CalendarMode = "day" | "month" | "year";

interface CaptionComponentProps extends React.HTMLAttributes<HTMLDivElement> {
  readonly calendarMonth?: { month?: Date; year?: number } | Date;
  readonly displayIndex?: number;
}

type CalendarDateRange = { from?: Date; to?: Date };
type CalendarSelected = Date | string | Date[] | CalendarDateRange;
type CalendarOnSelectValue = Date | Date[] | CalendarDateRange | undefined;

interface ChevronProps {
  readonly className?: string;
  readonly size?: number;
  readonly disabled?: boolean;
  readonly orientation?: "left" | "right" | "up" | "down";
}

function ChevronComponent(chevronProps: ChevronProps) {
  if (chevronProps.orientation === "left") {
    return <ChevronLeftIcon size={16} {...chevronProps} aria-hidden="true" />;
  }
  return <ChevronRightIcon size={16} {...chevronProps} aria-hidden="true" />;
}

function createLocalIsoString(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}T00:00:00.000Z`;
}

function toSerializableLocalDate(date: Date): Date {
  const normalized = new Date(
    date.getFullYear(),
    date.getMonth(),
    date.getDate(),
    0,
    0,
    0,
    0,
  );
  const localIso = createLocalIsoString(normalized);

  Object.defineProperty(normalized, "toISOString", {
    value: () => localIso,
    writable: true,
    configurable: true,
  });
  Object.defineProperty(normalized, "toJSON", {
    value: () => localIso,
    writable: true,
    configurable: true,
  });

  return normalized;
}

function normalizeSelectedValue(value: CalendarOnSelectValue): CalendarOnSelectValue {
  if (value instanceof Date) {
    return toSerializableLocalDate(value);
  }

  if (Array.isArray(value)) {
    return value.map((date) => toSerializableLocalDate(date));
  }

  if (value && typeof value === "object") {
    return {
      from: value.from ? toSerializableLocalDate(value.from) : undefined,
      to: value.to ? toSerializableLocalDate(value.to) : undefined,
    };
  }

  return value;
}

interface CaptionComponentFactoryProps {
  readonly months: readonly string[];
  readonly mergedClassNames: {
    readonly month_caption: string;
    readonly button_previous: string;
    readonly button_next: string;
  };
  readonly handlePreviousClick: () => void;
  readonly handleNextClick: () => void;
  readonly month: Date;
  readonly setMode: (mode: CalendarMode) => void;
}

function createCaptionComponentFactory({
  months,
  mergedClassNames,
  handlePreviousClick,
  handleNextClick,
  month,
  setMode,
}: CaptionComponentFactoryProps) {
  const CaptionComponent = (captionProps: CaptionComponentProps) => {
    // Extract only valid DOM props, exclude custom caption props
    const {
      calendarMonth,
      ...restProps
    } = captionProps;
    // Filter out any non-DOM props that might have leaked through
    const divProps = Object.fromEntries(
      Object.entries(restProps).filter(([key]) => {
        // Only allow valid HTML div attributes
        return !["calendarMonth", "displayIndex"].includes(key);
      }),
    ) as React.HTMLAttributes<HTMLDivElement>;
    // Handle different possible structures from react-day-picker
    // (v9+ passes { date: Date } while older versions may pass { month: Date }).
    let displayMonthProp: Date | undefined;
    if (calendarMonth instanceof Date) {
      displayMonthProp = calendarMonth;
    } else if (calendarMonth && typeof calendarMonth === "object") {
      if (
        "date" in calendarMonth &&
        calendarMonth.date instanceof Date
      ) {
        displayMonthProp = calendarMonth.date;
      } else if (
        "month" in calendarMonth &&
        calendarMonth.month instanceof Date
      ) {
        displayMonthProp = calendarMonth.month;
      }
    }

    // Use month state as fallback if displayMonthProp is not available
    const safeDisplayMonth = displayMonthProp ?? month;

    // Always show the same header UI regardless of mode
    return (
      <div {...divProps} className={mergedClassNames.month_caption}>
        <button
          type="button"
          onClick={handlePreviousClick}
          className={mergedClassNames.button_previous}
          aria-label="Previous month"
        >
          <ChevronLeftIcon size={16} aria-hidden="true" />
        </button>
        <div className="flex items-center justify-center gap-2">
          <button
            type="button"
            onClick={() => setMode("month")}
            className="text-sm font-medium text-primary hover:text-primary/80 transition-colors px-2 py-1 rounded-md hover:bg-accent cursor-pointer"
          >
            {months[safeDisplayMonth.getMonth()]}
          </button>
          <button
            type="button"
            onClick={() => setMode("year")}
            className="text-sm font-medium text-primary hover:text-primary/80 transition-colors px-2 py-1 rounded-md hover:bg-accent cursor-pointer"
          >
            {safeDisplayMonth.getFullYear()}
          </button>
        </div>
        <button
          type="button"
          onClick={handleNextClick}
          className={mergedClassNames.button_next}
          aria-label="Next month"
        >
          <ChevronRightIcon size={16} aria-hidden="true" />
        </button>
      </div>
    );
  };
  return CaptionComponent;
}

// Use a more permissive type to handle DayPicker's conditional types
type CalendarPropsBase = Omit<DayPickerProps, "month" | "required">;
type CalendarProps = CalendarPropsBase & {
  readonly className?: string;
  readonly classNames?: DayPickerProps["classNames"];
  readonly showOutsideDays?: boolean;
  readonly components?: DayPickerProps["components"];
  readonly month?: Date;
  readonly onMonthChange?: (month: Date | undefined) => void;
  // Explicitly allow common DayPicker props that are conditionally typed
  readonly mode?: "single" | "multiple" | "range";
  readonly selected?: CalendarSelected;
  readonly onSelect?:
    | ((date: Date | undefined) => void)
    | ((dates: Date[] | undefined) => void)
    | ((range: CalendarDateRange | undefined) => void);
};

function Calendar({
  className,
  classNames,
  showOutsideDays = true,
  components: userComponents,
  month: controlledMonth,
  onMonthChange: controlledOnMonthChange,
  ...props
}: CalendarProps) {
  // Destructure out props that conflict with DayPicker's conditional types
  const {
    required,
    selected: rawSelected,
    onSelect: rawOnSelect,
    ...safeProps
  } = props as CalendarProps & {
    required?: boolean;
    selected?: CalendarProps["selected"];
  };
  void required;
  const [internalMonth, setInternalMonth] = React.useState<Date>(
    controlledMonth ?? new Date(),
  );
  const [mode, setMode] = React.useState<CalendarMode>("day");

  const onMonthChange = controlledOnMonthChange ?? setInternalMonth;

  const normalizedSelected = React.useMemo(() => {
    if (typeof rawSelected === "string") {
      return parseLocalDate(rawSelected);
    }

    return rawSelected;
  }, [rawSelected]);

  const handleSelect = React.useCallback((value: CalendarOnSelectValue) => {
    (rawOnSelect as ((selected: CalendarOnSelectValue) => void) | undefined)?.(
      normalizeSelectedValue(value),
    );
  }, [rawOnSelect]);

  const displayMonth = controlledMonth ?? internalMonth;

  const defaultClassNames = {
    months: "relative flex flex-col sm:flex-row gap-4",
    month: "w-full",
    month_caption: "relative mb-1 flex h-9 items-center justify-between z-20",
    caption_label: "text-sm font-medium",
    nav: "hidden", // Hide default navigation since we use custom navigation in caption
    button_previous: cn(
      buttonVariants({ variant: "ghost" }),
      "size-9 p-0 text-muted-foreground/80 hover:text-foreground",
    ),
    button_next: cn(
      buttonVariants({ variant: "ghost" }),
      "size-9 p-0 text-muted-foreground/80 hover:text-foreground",
    ),
    weekday: "size-9 p-0 text-xs font-medium text-muted-foreground/80",
    day_button:
      "relative flex size-9 items-center justify-center whitespace-nowrap rounded-md p-0 text-foreground group-[[data-selected]:not(.range-middle)]:[transition-property:color,background-color,border-radius,box-shadow] group-[[data-selected]:not(.range-middle)]:duration-150 group-data-disabled:pointer-events-none focus-visible:z-10 hover:not-in-data-selected:bg-accent group-data-selected:bg-primary hover:not-in-data-selected:text-foreground group-data-selected:text-primary-foreground group-data-disabled:text-foreground/30 group-data-disabled:line-through group-data-outside:text-foreground/30 group-data-selected:group-data-outside:text-primary-foreground outline-none focus-visible:ring-ring/50 focus-visible:ring-[3px] group-[.range-start:not(.range-end)]:rounded-e-none group-[.range-end:not(.range-start)]:rounded-s-none group-[.range-middle]:rounded-none group-[.range-middle]:group-data-selected:bg-accent group-[.range-middle]:group-data-selected:text-foreground",
    day: "group size-9 px-0 py-px text-sm",
    range_start: "range-start",
    range_end: "range-end",
    range_middle: "range-middle",
    today:
      "*:after:pointer-events-none *:after:absolute *:after:bottom-1 *:after:start-1/2 *:after:z-10 *:after:size-[3px] *:after:-translate-x-1/2 *:after:rounded-full *:after:bg-primary [&[data-selected]:not(.range-middle)>*]:after:bg-background [&[data-disabled]>*]:after:bg-foreground/30 *:after:transition-colors",
    outside:
      "text-muted-foreground data-selected:bg-accent/50 data-selected:text-muted-foreground",
    hidden: "invisible",
    week_number: "size-9 p-0 text-xs font-medium text-muted-foreground/80",
  };

  const allClassNameKeys = new Set([
    ...Object.keys(defaultClassNames),
    ...Object.keys(classNames ?? {}),
  ]);
  const mergedClassNames = Array.from(allClassNameKeys).reduce(
    (acc, key) => {
      const baseClassName =
        defaultClassNames[key as keyof typeof defaultClassNames] ?? "";
      const userClassName = classNames?.[key as keyof typeof classNames];
      const shouldHide =
        mode !== "day" &&
        (key === "day" || key === "day_button" || key === "weekday");
      const hideClassName = shouldHide ? "hidden" : undefined;

      return {
        ...acc,
        [key]: userClassName
          ? cn(baseClassName, userClassName, hideClassName)
          : cn(baseClassName, hideClassName),
      };
    },
    {} as Record<string, string>,
  );

  const months = React.useMemo(
    () => [
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
    ],
    [],
  );

  const currentYear = displayMonth.getFullYear();
  const currentMonth = displayMonth.getMonth();
  const startYear = currentYear - 10;
  const endYear = currentYear + 10;
  const years = Array.from(
    { length: endYear - startYear + 1 },
    (_, i) => startYear + i,
  );

  const handleMonthSelect = React.useCallback(
    (selectedMonth: number) => {
      const newDate = new Date(displayMonth);
      newDate.setMonth(selectedMonth);
      onMonthChange(newDate);
      setMode("day");
    },
    [displayMonth, onMonthChange],
  );

  const handleYearSelect = React.useCallback(
    (selectedYear: number) => {
      const newDate = new Date(displayMonth);
      newDate.setFullYear(selectedYear);
      onMonthChange(newDate);
      setMode("day");
    },
    [displayMonth, onMonthChange],
  );

  const handlePreviousClick = React.useCallback(() => {
    if (mode === "month") {
      const newDate = new Date(displayMonth);
      newDate.setFullYear(newDate.getFullYear() - 1);
      onMonthChange(newDate);
    } else if (mode === "year") {
      const newDate = new Date(displayMonth);
      newDate.setFullYear(newDate.getFullYear() - 20);
      onMonthChange(newDate);
    } else {
      const newDate = new Date(displayMonth);
      newDate.setMonth(newDate.getMonth() - 1);
      onMonthChange(newDate);
    }
  }, [mode, displayMonth, onMonthChange]);

  const handleNextClick = React.useCallback(() => {
    if (mode === "month") {
      const newDate = new Date(displayMonth);
      newDate.setFullYear(newDate.getFullYear() + 1);
      onMonthChange(newDate);
    } else if (mode === "year") {
      const newDate = new Date(displayMonth);
      newDate.setFullYear(newDate.getFullYear() + 20);
      onMonthChange(newDate);
    } else {
      const newDate = new Date(displayMonth);
      newDate.setMonth(newDate.getMonth() + 1);
      onMonthChange(newDate);
    }
  }, [mode, displayMonth, onMonthChange]);

  const CustomCaption = React.useMemo(
    () =>
      createCaptionComponentFactory({
        months,
        mergedClassNames: {
          month_caption: mergedClassNames.month_caption,
          button_previous: mergedClassNames.button_previous,
          button_next: mergedClassNames.button_next,
        },
        handlePreviousClick,
        handleNextClick,
        month: displayMonth,
        setMode,
      }),
    [
      months,
      mergedClassNames,
      handlePreviousClick,
      handleNextClick,
      displayMonth,
      setMode,
    ],
  );

  const defaultComponents = React.useMemo(
    () => ({
      Chevron: ChevronComponent,
      MonthCaption: CustomCaption as NonNullable<
        DayPickerProps["components"]
      >["MonthCaption"],
      // Hide default navigation since we use custom navigation in caption
      Navigation: () => null,
    }),
    [CustomCaption],
  );

  const mergedComponents = {
    ...defaultComponents,
    ...userComponents,
  };

  // Custom content for month/year selection that appears below the header
  const monthYearContent = React.useMemo(() => {
    if (mode === "month") {
      return (
        <div className="grid grid-cols-3 gap-2 p-4 border-t">
          {months.map((monthName, index) => (
            <button
              key={monthName}
              type="button"
              onClick={() => handleMonthSelect(index)}
              className={cn(
                "px-4 py-2 text-sm rounded-md transition-colors",
                 index === currentMonth && currentYear === displayMonth.getFullYear()
                  ? "bg-primary text-primary-foreground font-medium"
                  : "hover:bg-accent hover:text-accent-foreground",
              )}
            >
              {monthName.slice(0, 3)}
            </button>
          ))}
        </div>
      );
    }

    if (mode === "year") {
      return (
        <div className="grid grid-cols-4 gap-2 p-4 border-t max-h-[300px] overflow-y-auto">
          {years.map((year) => (
            <button
              key={year}
              type="button"
              onClick={() => handleYearSelect(year)}
              className={cn(
                "px-4 py-2 text-sm rounded-md transition-colors",
                year === currentYear
                  ? "bg-primary text-primary-foreground font-medium"
                  : "hover:bg-accent hover:text-accent-foreground",
              )}
            >
              {year}
            </button>
          ))}
        </div>
      );
    }

    return null;
  }, [
    mode,
    months,
    years,
    currentMonth,
    currentYear,
    displayMonth,
    handleMonthSelect,
    handleYearSelect,
  ]);

  return (
    <div className={cn(className)}>
      <DayPicker
        {...({
          ...safeProps,
          showOutsideDays: mode === "day" ? showOutsideDays : false,
          className: cn("w-full", className),
          classNames: mergedClassNames,
          components: mergedComponents,
          month: displayMonth,
          onMonthChange,
          selected: normalizedSelected,
          onSelect: rawOnSelect ? handleSelect : undefined,
        } as DayPickerProps)}
      />
      {monthYearContent}
    </div>
  );
}

export { Calendar };
export type { DateRange } from "react-day-picker";
