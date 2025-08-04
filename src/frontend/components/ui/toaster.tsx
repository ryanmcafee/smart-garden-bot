'use client';

import * as React from 'react';

// Simple toaster placeholder component
export function Toaster() {
  return (
    <div 
      id="toaster"
      className="fixed bottom-4 right-4 z-50 pointer-events-none"
      aria-live="polite"
      aria-label="Notifications"
    />
  );
}