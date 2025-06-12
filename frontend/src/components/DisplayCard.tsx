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
  tags // <-- destructure tags from props
}) => {
  const defaultImage = type === 'Recipe' ? '/recipe-blank.jpg' : '/meal-blank.jpg';
  const currentImageUrl = imageUrl || defaultImage;
  const altText = imageAltText || `${title || type} image`;

  const handleImageError = (e: React.SyntheticEvent<HTMLImageElement, Event>) => {
    (e.target as HTMLImageElement).src = defaultImage;
  };

  return (
    <div key={id} className="card bg-base-100 shadow-sm hover:shadow-md transition-shadow duration-300 ease-in-out transform hover:-translate-y-1">
      <figure>
        <img
          src={currentImageUrl}
          alt={altText}
          className="w-full h-full object-cover"
          onError={handleImageError}
        />
      </figure>
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
              <span key={tag} className="badge badge-primary badge-sm">{tag}</span>
            ))}
          </div>
        )}
        {description && (
          <p className="text-sm text-gray-600 mb-4 h-20 overflow-hidden text-ellipsis" title={description}>
            {description.length > 120 ? `${description.substring(0, 117)}...` : description}
          </p>
        )}
        {!description && <p className="text-sm text-gray-500 mb-4 h-20 italic">No description available.</p>}
        <div className="card-actions justify-end mt-auto">
          <Link to={viewLink} className="btn btn-sm btn-outline btn-primary">
            View {type}
          </Link>
          {editLink && (
            <Link to={editLink} className="btn btn-sm btn-outline ml-2">
              Edit
            </Link>
          )}
        </div>
      </div>
    </div>
  );
};

export default DisplayCard;
