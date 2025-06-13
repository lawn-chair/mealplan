// filepath: /Users/brent/code/mealplan/frontend/src/components/DisplayCard.tsx
import React from 'react';
import { Link } from 'react-router-dom';

interface DisplayCardProps {
  id: string | number | undefined; // Changed to string | number | undefined for flexibility
  imageUrl?: string;
  title: string;
  description?: string;
  viewLink: string;
  editLink?: string; // Optional: for an edit button
  imageAltText?: string;
  type?: 'Recipe' | 'Meal' | 'Item'; // Added 'Item' to match default
  tags?: string[]; // Optional: tags to display as badges
  onTagClick?: (tag: string) => void; // Optional: callback for tag click
  onAddToPlan?: () => void; // Optional: handler for add to plan (for meals)
}

const DisplayCard: React.FC<DisplayCardProps> = ({
  id,
  imageUrl,
  title,
  description,
  viewLink,
  editLink,
  imageAltText,
  type = 'Item',
  tags, // <-- destructure tags from props
  onTagClick,
  onAddToPlan
}) => {
  const defaultImage = type === 'Recipe' ? '/recipe-blank.jpg' : '/meal-blank.jpg';

  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    (e.target as HTMLImageElement).src = defaultImage;
  };

  return (
    <div key={id} className="card bg-base-100 shadow-sm hover:shadow-md transition-shadow duration-300 ease-in-out transform hover:-translate-y-1">
      {imageUrl && (
        <figure className="w-full h-48 bg-base-200 flex items-center justify-center overflow-hidden rounded-t-lg">
          <img
            src={imageUrl}
            alt={imageAltText || title}
            className="object-cover w-full h-full"
            loading="lazy"
            onError={handleImageError}
          />
        </figure>
      )}
      <div className="card-body p-6">
        <h3 className="card-title text-xl font-semibold mb-2 whitespace-normal break-words" title={title}>
          <Link to={viewLink} className="link link-hover link-primary">
            {title || `Untitled ${type}`}
          </Link>
        </h3>
        {/* Tag display */}
        {tags && tags.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2">
            {tags.map((tag: string) => (
              onTagClick ? (
                <button
                  key={tag}
                  className="badge badge-primary badge-sm cursor-pointer hover:badge-accent transition-colors"
                  type="button"
                  onClick={() => onTagClick(tag)}
                  tabIndex={0}
                  aria-label={`Filter by tag ${tag}`}
                >
                  {tag}
                </button>
              ) : (
                <span key={tag} className="badge badge-primary badge-sm">{tag}</span>
              )
            ))}
          </div>
        )}
        {description && (
          <p className="text-sm text-gray-600 mb-4 h-20 overflow-hidden text-ellipsis" title={description}>
            {description.length > 120 ? `${description.substring(0, 117)}...` : description}
          </p>
        )}
        {!description && <p className="text-sm text-gray-500 mb-4 h-20 italic">No description available.</p>}
        <div className="card-actions justify-end mt-auto gap-2">
          <Link to={viewLink} className="btn btn-sm btn-outline btn-primary">
            View {type}
          </Link>
          {editLink && (
            <Link to={editLink} className="btn btn-sm btn-outline ml-2">
              Edit
            </Link>
          )}
          {onAddToPlan && (
            <button
              className="btn btn-sm btn-outline btn-primary"
              onClick={onAddToPlan}
              type="button"
            >
              + Add to Plan
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default DisplayCard;
