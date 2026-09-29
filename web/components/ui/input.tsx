import { forwardRef } from "react";
import { clsx } from "clsx";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  hint?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, hint, className, id, ...props }, ref) => {
    const inputId = id ?? label?.toLowerCase().replace(/\s+/g, "-");
    return (
      <div className="w-full">
        {label && (
          <label htmlFor={inputId} className="mb-1.5 block text-sm font-medium text-ink">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          className={clsx(
            "w-full rounded-lg border bg-card px-4 py-2.5 text-sm text-ink placeholder-muted/70",
            "transition-colors focus:outline-none focus:ring-2 focus:ring-primary-500/40",
            error ? "border-red-400 focus:ring-red-400/40" : "border-border hover:border-muted/50",
            className,
          )}
          {...props}
        />
        {error && <p className="mt-1.5 text-xs font-medium text-red-500">{error}</p>}
        {hint && !error && <p className="mt-1.5 text-xs text-muted">{hint}</p>}
      </div>
    );
  },
);
Input.displayName = "Input";
