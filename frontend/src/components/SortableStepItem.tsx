import { ChangeEvent } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';

export interface CommonStep {
  id: string | number; // This was the original, let's ensure it's used correctly or make it flexible if needed.
  text: string;
  // Potentially add order here if it's truly common, but it seems specific to the form's state management
}

export interface SortableStepItemProps<T extends CommonStep> {
  item: T;
  index: number; // Visual index for display (e.g., "Step 1")
  onTextChange: (itemId: string | number, newText: string) => void;
  onRemove: (itemId: string | number) => void;
  isDraggable?: boolean;
  canRemove?: boolean;
}

export function SortableStepItem<T extends CommonStep>({
  item,
  index,
  onTextChange,
  onRemove,
  isDraggable = true,
  canRemove = true,
}: SortableStepItemProps<T>) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ id: item.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    zIndex: isDragging ? 100 : 'auto',
    opacity: isDragging ? 0.8 : 1,
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className="flex items-start gap-2 p-3 bg-base-200 rounded shadow-sm mb-2"
    >
      {isDraggable && (
        <button
          type="button"
          {...attributes}
          {...listeners}
          className="btn btn-xs btn-ghost cursor-grab p-1 mt-1" // Added mt-1 for alignment with textarea
          aria-label="Drag to reorder step"
        >
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-5 h-5">
            <path strokeLinecap="round" strokeLinejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
          </svg>
        </button>
      )}
      {!isDraggable && <div className="w-8 h-8"></div>} {/* Placeholder for alignment when not draggable */}
      
      <span className="text-sm font-semibold mr-2 p-1 mt-1">{index + 1}.</span>
      <textarea
        placeholder="Describe this step..."
        value={item.text}
        onChange={(e: ChangeEvent<HTMLTextAreaElement>) => onTextChange(item.id, e.target.value)}
        required
        className="textarea textarea-bordered textarea-primary textarea-sm flex-grow h-24"
      />
      {canRemove && (
        <button
          type="button"
          onClick={() => onRemove(item.id)}
          className="btn btn-sm btn-outline btn-error mt-1" // Added mt-1 for alignment
        >
          <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-1" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
          Remove
        </button>
      )}
       {!canRemove && <div className="w-[90px]"></div>} {/* Placeholder for alignment */}
    </div>
  );
}
