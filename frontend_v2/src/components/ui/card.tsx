import { cn } from "@/lib/utils";

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
}

export function Card({ children, className, ...props }: CardProps) {
  return (
    <div 
      className={cn("bg-card border border-border rounded-[2rem] p-6 shadow-premium", className)} 
      {...props}
    >
      {children}
    </div>
  );
}
