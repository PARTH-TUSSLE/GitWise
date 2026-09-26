import { Request, Response } from 'express';
import * as path from 'path';
const fs = require('fs');

export interface UserProfile {
  id: string;
  username: string;
  email?: string;
}

export type AccountStatus = 'active' | 'suspended' | 'pending';

export class UserService {
  private cache: Map<string, UserProfile> = new Map();

  public async getUser(id: string): Promise<UserProfile | null> {
    if (this.cache.has(id)) {
      return this.cache.get(id) || null;
    }
    return null;
  }

  public clear(): void {
    this.cache.clear();
  }
}

export async function fetchRemoteUser(apiUrl: string): Promise<UserProfile> {
  return { id: '1', username: 'alice' };
}

export const formatUserName = (user: UserProfile): string => {
  return user.username.toUpperCase();
};

const internalLocalHelper = (x: number): number => {
  return x + 1;
};

export default function defaultHandler(req: Request, res: Response): void {
  res.send('ok');
}
