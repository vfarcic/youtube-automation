interface FieldLabelProps {
  name: string;
  required?: boolean;
  helpText?: string;
  htmlFor?: string;
  complete?: boolean;
}

export function FieldLabel({ name, required, helpText, htmlFor, complete }: FieldLabelProps) {
  return (
    <div className="mb-1">
      <label htmlFor={htmlFor} className="flex items-center gap-1.5 text-sm font-medium text-gray-300">
        {complete !== undefined && !complete && (
          <span
            className="inline-block w-2 h-2 rounded-full shrink-0 bg-red-400"
          />
        )}
        {name}
        {required && <span className="text-red-500 ml-0.5">*</span>}
      </label>
      {helpText && (
        <p className="text-xs text-gray-500 mt-0.5">{helpText}</p>
      )}
    </div>
  );
}
