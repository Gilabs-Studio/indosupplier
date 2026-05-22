import * as React from "react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";

interface NumericInputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "onChange" | "value"> {
  value?: number | string;
  onChange?: (value: number | undefined) => void;
  currency?: boolean;
}

const NumericInput = React.forwardRef<HTMLInputElement, NumericInputProps>(
  ({ className, value, onChange, ...props }, ref) => {
    // Internal state for the display value
    const [displayValue, setDisplayValue] = React.useState("");

    // Format number to Indonesian locale string
    const formatNumber = (num: number | undefined | string) => {
      if (num === undefined || num === "" || num === null) return "";
      const n = typeof num === 'string' ? parseFloat(num) : num;
      if (isNaN(n)) return "";
      
      return new Intl.NumberFormat("id-ID", {
        maximumFractionDigits: 20, // Allow high precision input
      }).format(n);
    };

    // Initialize display value from props
    React.useEffect(() => {
        if (value !== undefined && value !== null) {
            // Only update display value from prop if it doesn't match the current parsed display value
            // This prevents cursor jumping / reformatting loop issues during typing if we were strict
            // But for simple "controlled" input, we usually want to sync.
            // For now, let's sync.
            const formatted = formatNumber(value);
             // Verify if current display value is roughly equivalent to avoid resetting "1000." to "1000" while typing decimal
             const cleanDisplay = displayValue.replace(/\./g, "").replace(",", ".");
             const cleanValue = value.toString();
             if (parseFloat(cleanDisplay) !== parseFloat(cleanValue)) {
                 setDisplayValue(formatted);
             } else if (displayValue === "") {
                 setDisplayValue(formatted);
             }
        } else {
            if (value === undefined) setDisplayValue("");
        }
    }, [value, displayValue]);


    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const inputValue = e.target.value;
      
      // Allow only numbers, commas, and dots
      // In ID locale, dot is grouping, comma is decimal
      // We accept input, remove dots, replace comma with dot for parsing
      
      const cleanValue = inputValue.replace(/[^0-9,]/g, ""); // Remove non-numeric except comma
      
      // Handle display update (let user type freely)
      // We might want to add formatting on the fly, but it can be annoying with cursor.
      // Better approach: 
      // 1. Let user type.
      // 2. Parse effective number.
      // 3. Call onChange with effective number.
      // 4. Update displayValue to formatted version ONLY on blur or if we are confident?
      // Actually, user wants "10000 -> 10.000". This usually implies real-time formatting.
      
      // Let's try simple real-time formatting for integer part, careful with decimals.
      
      const parts = cleanValue.split(",");
      const integerPart = parts[0].replace(/\./g, ""); // remove existing dots to re-format
      const decimalPart = parts[1];
      
      // Format integer part
      const formattedInteger = integerPart.replace(/\B(?=(\d{3})+(?!\d))/g, ".");
      
      let newDisplayValue = formattedInteger;
      if (decimalPart !== undefined) {
         newDisplayValue += "," + decimalPart;
      } else if (inputValue.endsWith(",")) {
         newDisplayValue += ",";
      }

      setDisplayValue(newDisplayValue);

      // Parse for parent
      const parseableString = cleanValue.replace(/\./g, "").replace(",", ".");
      const numberValue = parseableString ? parseFloat(parseableString) : undefined;
      
      onChange?.(numberValue);
    };

    const handleBlur = () => {
        // On blur, ensure it is strictly formatted
        if (value !== undefined) {
            setDisplayValue(formatNumber(value));
        }
    }

    return (
      <Input
        type="text" // Must be text to support formatted strings
        inputMode="decimal"
        className={cn(className)}
        value={displayValue}
        onChange={handleChange}
        onBlur={handleBlur}
        ref={ref}
        {...props}
      />
    );
  }
);

NumericInput.displayName = "NumericInput";

export { NumericInput };
