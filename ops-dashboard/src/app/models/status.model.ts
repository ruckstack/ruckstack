export interface StatusModel {
  healthy: boolean;
  name: string;
  support: string[];
  version: string;
  buildTime: number;
  buildDate: Date;
  level: number;
}

export interface StatusDetailedModel extends StatusModel {
  trackers: Tracker[]
}

export interface Tracker {
  Name: string
  CurrentProblems: Map<string, string>
  CurrentWarnings: Map<string, string>
}
