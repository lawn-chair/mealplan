import React, { useEffect } from 'react';

interface PageTitleUpdaterProps {
  section: string;
}

const PageTitleUpdater: React.FC<PageTitleUpdaterProps> = ({ section }) => {
  useEffect(() => {
    if (section) {
      document.title = `Yum - ${section}`;
    } else {
      document.title = 'Yum - Plan Your Meals'; // Default title
    }
    // Optional: Reset title on component unmount if needed
    // return () => { document.title = 'Yum - Plan Your Meals'; };
  }, [section]);

  return null; // This component does not render anything
};

export default PageTitleUpdater;
