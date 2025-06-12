import React, { useState, KeyboardEvent, ChangeEvent, useEffect } from 'react';
import { getTags } from '../api';

interface TagInputProps {
  tags: string[];
  setTags: (tags: string[]) => void;
  label?: string;
  placeholder?: string;
  disabled?: boolean;
}

const TagInput: React.FC<TagInputProps> = ({ tags, setTags, label = 'Tags', placeholder = 'Add a tag and press Enter', disabled }) => {
  const [input, setInput] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [filteredSuggestions, setFilteredSuggestions] = useState<string[]>([]);
  useEffect(() => {
    // Fetch all tags for autocomplete using the new API helper
    getTags().then(res => {
      setSuggestions(res.data || []);
    }).catch(() => setSuggestions([]));
  }, []);

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
    setShowSuggestions(true);
  };

  const handleInputKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if ((e.key === 'Enter' || e.key === ',') && input.trim()) {
      e.preventDefault();
      const newTag = input.trim().toLowerCase();
      if (newTag && !tags.includes(newTag)) {
        setTags([...tags, newTag]);
      }
      setInput('');
      setShowSuggestions(false);
    } else if (e.key === 'Backspace' && !input && tags.length > 0) {
      setTags(tags.slice(0, -1));
    }
  };

  const removeTag = (index: number) => {
    setTags(tags.filter((_, i) => i !== index));
  };

  const handleSuggestionClick = (suggestion: string) => {
    if (!tags.includes(suggestion)) {
      setTags([...tags, suggestion]);
    }
    setInput('');
    setShowSuggestions(false);
  };

  useEffect(() => {
    console.log('Filtering suggestions based on input:', input, "suggestions: ", suggestions);
      if (suggestions) {
          setFilteredSuggestions(
              suggestions.filter(
                  (s) => s.includes(input.trim().toLowerCase()) && !tags.includes(s)
              ).slice(0, 8)
          );
      }
  }, [suggestions, input, tags]);


  return (
    <div className="form-control">
      {label && <label className="label"><span className="label-text">{label}</span></label>}
      <div className="flex flex-wrap gap-2 items-center bg-base-200 rounded p-2 min-h-[44px] relative">
        {tags.map((tag, idx) => (
          <span key={tag} className="badge badge-primary badge-md flex items-center gap-1">
            {tag}
            {!disabled && (
              <button
                type="button"
                className="btn btn-xs btn-circle btn-ghost ml-1"
                aria-label={`Remove tag ${tag}`}
                onClick={() => removeTag(idx)}
                tabIndex={-1}
              >
                ×
              </button>
            )}
          </span>
        ))}
        <div className="relative flex-1 min-w-[100px]">
          <input
            type="text"
            className="input input-bordered input-sm w-full bg-transparent focus:bg-base-100"
            value={input}
            onChange={handleInputChange}
            onKeyDown={handleInputKeyDown}
            placeholder={placeholder}
            disabled={disabled}
            onFocus={() => setShowSuggestions(true)}
            onBlur={() => setTimeout(() => setShowSuggestions(false), 100)}
            aria-label="Add tag"
          />
          {showSuggestions && input && filteredSuggestions.length > 0 && (
            <ul className="absolute z-10 left-0 right-0 bg-base-100 border border-base-300 rounded shadow mt-1 max-h-40 overflow-y-auto">
              {filteredSuggestions.map((s) => (
                <li
                  key={s}
                  className="px-3 py-2 cursor-pointer hover:bg-base-200"
                  onMouseDown={() => handleSuggestionClick(s)}
                >
                  {s}
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default TagInput;
