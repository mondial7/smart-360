/**
 * User Role Context
 *
 * Provides user role information and helper methods
 */

import React, { createContext, useContext, ReactNode } from 'react';
import { useAuth } from './AuthContext';
import { UserRole } from '../types';

interface UserRoleContextType {
  role: UserRole | null;
  isAdmin: () => boolean;
  isMember: () => boolean;
}

const UserRoleContext = createContext<UserRoleContextType | undefined>(undefined);

export const useUserRole = () => {
  const context = useContext(UserRoleContext);
  if (context === undefined) {
    throw new Error('useUserRole must be used within a UserRoleProvider');
  }
  return context;
};

interface UserRoleProviderProps {
  children: ReactNode;
}

export const UserRoleProvider: React.FC<UserRoleProviderProps> = ({ children }) => {
  const { userProfile } = useAuth();

  const isAdmin = (): boolean => {
    return userProfile?.role === 'admin';
  };

  const isMember = (): boolean => {
    return userProfile?.role === 'member';
  };

  const value: UserRoleContextType = {
    role: userProfile?.role || null,
    isAdmin,
    isMember,
  };

  return <UserRoleContext.Provider value={value}>{children}</UserRoleContext.Provider>;
};
