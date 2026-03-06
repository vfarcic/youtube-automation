import { FieldLabel } from './FieldLabel';
import type { SelectOption } from '../../api/types';

interface SelectInputProps {
  name: string;
  fieldName: string;
  value: string;
  onChange: (fieldName: string, value: string) => void;
  options: SelectOption[];
  required?: boolean;
  helpText?: string;
  placeholder?: string;
  complete?: boolean;
}

export function SelectInput({
  name,
  fieldName,
  value,
  onChange,
  options,
  required,
  helpText,
  placeholder,
  complete,
}: SelectInputProps) {
  return (
    <div>
      <FieldLabel name={name} required={required} helpText={helpText} htmlFor={fieldName} complete={complete} />
      <select
        id={fieldName}
        value={value}
        onChange={(e) => onChange(fieldName, e.target.value)}
        className="w-full border border-gray-600 bg-gray-800 text-gray-100 rounded px-3 py-1.5 text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      >
        {placeholder && <option value="">{placeholder}</option>}
        {options.map((opt) => (
          <option key={String(opt.value)} value={String(opt.value)}>
            {opt.label}
          </option>
        ))}
      </select>
    </div>
  );
}
