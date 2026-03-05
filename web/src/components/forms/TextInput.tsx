import { FieldLabel } from './FieldLabel';

interface TextInputProps {
  name: string;
  fieldName: string;
  value: string;
  onChange: (fieldName: string, value: string) => void;
  placeholder?: string;
  required?: boolean;
  helpText?: string;
  complete?: boolean;
}

export function TextInput({
  name,
  fieldName,
  value,
  onChange,
  placeholder,
  required,
  helpText,
  complete,
}: TextInputProps) {
  return (
    <div>
      <FieldLabel name={name} required={required} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <input
        id={fieldName}
        type="text"
        value={value}
        onChange={(e) => onChange(fieldName, e.target.value)}
        placeholder={placeholder}
        className="w-full border border-gray-300 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    </div>
  );
}
