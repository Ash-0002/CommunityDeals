// Simple Button component

import { ButtonHTMLAttributes, forwardRef } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  size?: "sm" | "md";
  variant?: "primary" | "outline";
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className = "", size = "md", variant = "primary", children, ...props }, ref) => {
    const base = "inline-flex items-center justify-center font-bold rounded-lg transition-colors disabled:opacity-50";
    const sizes = { sm: "px-3.5 py-1.5 text-xs", md: "px-4 py-2 text-sm" };
    const variants = {
      primary: "bg-primary-600 text-white hover:bg-primary-700",
      outline: "border border-border text-ink hover:bg-surface",
    };
    return (
      <button ref={ref} className={`${base} ${sizes[size]} ${variants[variant]} ${className}`} {...props}>
        {children}
      </button>
    );
  }
);
Button.displayName = "Button";
