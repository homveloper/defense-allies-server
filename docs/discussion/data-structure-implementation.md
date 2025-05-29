# Defense Allies 데이터 구조 구현

## 📋 문서 정보
- **작성일**: 2024년
- **버전**: v1.0
- **목적**: JSON Schema 기반 데이터 구조 및 Redis 연동 구현
- **기반**: [매트릭스 밸런싱 시스템](matrix-balancing-system.md)

## 🗄️ JSON Schema 정의

### 종족 데이터 스키마
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Race Data Schema",
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-z_]+$",
      "description": "종족 고유 식별자"
    },
    "name": {
      "type": "string",
      "description": "종족 표시 이름"
    },
    "theme": {
      "type": "string",
      "description": "종족 테마 설명"
    },
    "power_matrix": {
      "type": "array",
      "items": {
        "type": "array",
        "items": {
          "type": "number",
          "minimum": 0.1,
          "maximum": 2.0
        },
        "minItems": 2,
        "maxItems": 2
      },
      "minItems": 2,
      "maxItems": 2,
      "description": "2x2 파워 매트릭스"
    },
    "matrix_properties": {
      "type": "object",
      "properties": {
        "frobenius_norm": {
          "type": "number",
          "minimum": 1.9,
          "maximum": 2.1
        },
        "determinant": {
          "type": "number",
          "minimum": 0.0,
          "maximum": 2.0
        },
        "trace": {
          "type": "number",
          "minimum": 1.5,
          "maximum": 2.5
        },
        "eigenvalues": {
          "type": "array",
          "items": {
            "type": "number"
          },
          "minItems": 2,
          "maxItems": 2
        }
      },
      "required": ["frobenius_norm", "determinant", "trace", "eigenvalues"]
    },
    "towers": {
      "type": "object",
      "properties": {
        "basic": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "minItems": 3,
          "maxItems": 3
        },
        "advanced": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "minItems": 3,
          "maxItems": 3
        },
        "cooperation": {
          "type": "array",
          "items": {
            "type": "string"
          },
          "minItems": 3,
          "maxItems": 3
        }
      },
      "required": ["basic", "advanced", "cooperation"]
    },
    "environment_modifiers": {
      "type": "object",
      "properties": {
        "time": {
          "type": "object",
          "patternProperties": {
            "^(dawn|day|dusk|night)$": {
              "$ref": "#/definitions/matrix_2x2"
            }
          }
        },
        "weather": {
          "type": "object",
          "patternProperties": {
            "^(clear|rain|storm|snow|fog)$": {
              "$ref": "#/definitions/matrix_2x2"
            }
          }
        },
        "terrain": {
          "type": "object",
          "patternProperties": {
            "^(plain|forest|mountain|desert|swamp|urban)$": {
              "$ref": "#/definitions/matrix_2x2"
            }
          }
        }
      },
      "required": ["time", "weather", "terrain"]
    },
    "synergy_coefficients": {
      "type": "object",
      "patternProperties": {
        "^[a-z_]+$": {
          "type": "number",
          "minimum": 0.1,
          "maximum": 2.0
        }
      },
      "description": "다른 종족과의 시너지 계수"
    }
  },
  "required": ["id", "name", "theme", "power_matrix", "matrix_properties", "towers", "environment_modifiers", "synergy_coefficients"],
  "definitions": {
    "matrix_2x2": {
      "type": "array",
      "items": {
        "type": "array",
        "items": {
          "type": "number",
          "minimum": 0.1,
          "maximum": 2.0
        },
        "minItems": 2,
        "maxItems": 2
      },
      "minItems": 2,
      "maxItems": 2
    }
  }
}
```

### 타워 데이터 스키마
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Tower Data Schema",
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "pattern": "^[a-z_]+$"
    },
    "name": {
      "type": "string"
    },
    "race_id": {
      "type": "string",
      "pattern": "^[a-z_]+$"
    },
    "tier": {
      "type": "string",
      "enum": ["basic", "advanced", "cooperation"]
    },
    "power_matrix": {
      "$ref": "#/definitions/matrix_2x2"
    },
    "cost": {
      "type": "object",
      "properties": {
        "gold": {"type": "integer", "minimum": 0},
        "mana": {"type": "integer", "minimum": 0},
        "special_resource": {"type": "integer", "minimum": 0}
      },
      "required": ["gold", "mana"]
    },
    "build_requirements": {
      "type": "object",
      "properties": {
        "prerequisite_towers": {
          "type": "array",
          "items": {"type": "string"}
        },
        "cooperation_players": {
          "type": "integer",
          "minimum": 1,
          "maximum": 4
        }
      }
    },
    "abilities": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "id": {"type": "string"},
          "name": {"type": "string"},
          "effect_matrix": {"$ref": "#/definitions/matrix_2x2"},
          "cooldown": {"type": "number", "minimum": 0},
          "range": {"type": "number", "minimum": 0}
        },
        "required": ["id", "name", "effect_matrix"]
      }
    }
  },
  "required": ["id", "name", "race_id", "tier", "power_matrix", "cost"],
  "definitions": {
    "matrix_2x2": {
      "type": "array",
      "items": {
        "type": "array",
        "items": {"type": "number"},
        "minItems": 2,
        "maxItems": 2
      },
      "minItems": 2,
      "maxItems": 2
    }
  }
}
```

## 🔧 Python 데이터 클래스 구현

### 종족 데이터 클래스
```python
from dataclasses import dataclass, field
from typing import Dict, List, Optional
import numpy as np
import json
import jsonschema

@dataclass
class MatrixProperties:
    """매트릭스 수학적 속성"""
    frobenius_norm: float
    determinant: float
    trace: float
    eigenvalues: List[float]
    
    def validate(self) -> bool:
        """매트릭스 속성 유효성 검증"""
        return (
            1.9 <= self.frobenius_norm <= 2.1 and
            0.0 <= self.determinant <= 2.0 and
            1.5 <= self.trace <= 2.5 and
            len(self.eigenvalues) == 2
        )

@dataclass
class TowerSet:
    """종족별 타워 세트"""
    basic: List[str] = field(default_factory=list)
    advanced: List[str] = field(default_factory=list)
    cooperation: List[str] = field(default_factory=list)
    
    def validate(self) -> bool:
        """타워 세트 유효성 검증"""
        return (
            len(self.basic) == 3 and
            len(self.advanced) == 3 and
            len(self.cooperation) == 3
        )

@dataclass
class EnvironmentModifiers:
    """환경 수정자"""
    time: Dict[str, np.ndarray] = field(default_factory=dict)
    weather: Dict[str, np.ndarray] = field(default_factory=dict)
    terrain: Dict[str, np.ndarray] = field(default_factory=dict)
    
    def get_combined_modifier(self, time: str, weather: str, terrain: str) -> np.ndarray:
        """복합 환경 수정자 계산"""
        result = np.array([[1.0, 1.0], [1.0, 1.0]])
        
        if time in self.time:
            result = result * self.time[time]
        if weather in self.weather:
            result = result * self.weather[weather]
        if terrain in self.terrain:
            result = result * self.terrain[terrain]
            
        return result

@dataclass
class RaceData:
    """종족 데이터"""
    id: str
    name: str
    theme: str
    power_matrix: np.ndarray
    matrix_properties: MatrixProperties
    towers: TowerSet
    environment_modifiers: EnvironmentModifiers
    synergy_coefficients: Dict[str, float] = field(default_factory=dict)
    
    def __post_init__(self):
        """초기화 후 검증"""
        self.validate()
    
    def validate(self) -> bool:
        """종족 데이터 유효성 검증"""
        if not self.matrix_properties.validate():
            raise ValueError(f"Invalid matrix properties for race {self.id}")
        
        if not self.towers.validate():
            raise ValueError(f"Invalid tower set for race {self.id}")
        
        if self.power_matrix.shape != (2, 2):
            raise ValueError(f"Power matrix must be 2x2 for race {self.id}")
        
        return True
    
    def to_dict(self) -> Dict:
        """딕셔너리로 변환"""
        return {
            'id': self.id,
            'name': self.name,
            'theme': self.theme,
            'power_matrix': self.power_matrix.tolist(),
            'matrix_properties': {
                'frobenius_norm': self.matrix_properties.frobenius_norm,
                'determinant': self.matrix_properties.determinant,
                'trace': self.matrix_properties.trace,
                'eigenvalues': self.matrix_properties.eigenvalues
            },
            'towers': {
                'basic': self.towers.basic,
                'advanced': self.towers.advanced,
                'cooperation': self.towers.cooperation
            },
            'environment_modifiers': {
                'time': {k: v.tolist() for k, v in self.environment_modifiers.time.items()},
                'weather': {k: v.tolist() for k, v in self.environment_modifiers.weather.items()},
                'terrain': {k: v.tolist() for k, v in self.environment_modifiers.terrain.items()}
            },
            'synergy_coefficients': self.synergy_coefficients
        }
    
    @classmethod
    def from_dict(cls, data: Dict) -> 'RaceData':
        """딕셔너리에서 생성"""
        matrix_props = MatrixProperties(**data['matrix_properties'])
        towers = TowerSet(**data['towers'])
        
        env_modifiers = EnvironmentModifiers()
        env_modifiers.time = {k: np.array(v) for k, v in data['environment_modifiers']['time'].items()}
        env_modifiers.weather = {k: np.array(v) for k, v in data['environment_modifiers']['weather'].items()}
        env_modifiers.terrain = {k: np.array(v) for k, v in data['environment_modifiers']['terrain'].items()}
        
        return cls(
            id=data['id'],
            name=data['name'],
            theme=data['theme'],
            power_matrix=np.array(data['power_matrix']),
            matrix_properties=matrix_props,
            towers=towers,
            environment_modifiers=env_modifiers,
            synergy_coefficients=data.get('synergy_coefficients', {})
        )

@dataclass
class TowerData:
    """타워 데이터"""
    id: str
    name: str
    race_id: str
    tier: str
    power_matrix: np.ndarray
    cost: Dict[str, int]
    build_requirements: Dict = field(default_factory=dict)
    abilities: List[Dict] = field(default_factory=list)
    
    def validate(self) -> bool:
        """타워 데이터 유효성 검증"""
        if self.tier not in ['basic', 'advanced', 'cooperation']:
            raise ValueError(f"Invalid tier {self.tier} for tower {self.id}")
        
        if self.power_matrix.shape != (2, 2):
            raise ValueError(f"Power matrix must be 2x2 for tower {self.id}")
        
        required_cost_fields = ['gold', 'mana']
        if not all(field in self.cost for field in required_cost_fields):
            raise ValueError(f"Missing required cost fields for tower {self.id}")
        
        return True
    
    def to_dict(self) -> Dict:
        """딕셔너리로 변환"""
        return {
            'id': self.id,
            'name': self.name,
            'race_id': self.race_id,
            'tier': self.tier,
            'power_matrix': self.power_matrix.tolist(),
            'cost': self.cost,
            'build_requirements': self.build_requirements,
            'abilities': self.abilities
        }
    
    @classmethod
    def from_dict(cls, data: Dict) -> 'TowerData':
        """딕셔너리에서 생성"""
        return cls(
            id=data['id'],
            name=data['name'],
            race_id=data['race_id'],
            tier=data['tier'],
            power_matrix=np.array(data['power_matrix']),
            cost=data['cost'],
            build_requirements=data.get('build_requirements', {}),
            abilities=data.get('abilities', [])
        )
```

## 🗃️ Redis 데이터 저장소 구현

### Redis 연동 클래스
```python
import redis
import json
import pickle
from typing import Optional, List, Dict, Any
import numpy as np

class DefenseAlliesRedisStore:
    """Defense Allies Redis 데이터 저장소"""
    
    def __init__(self, host: str = 'localhost', port: int = 6379, db: int = 0):
        self.redis_client = redis.Redis(host=host, port=port, db=db, decode_responses=True)
        self.binary_client = redis.Redis(host=host, port=port, db=db, decode_responses=False)
        
    # 종족 데이터 관리
    def save_race(self, race: RaceData) -> bool:
        """종족 데이터 저장"""
        try:
            key = f"race:{race.id}"
            data = race.to_dict()
            
            # JSON으로 저장 (읽기 쉬움)
            json_data = json.dumps(data, ensure_ascii=False)
            self.redis_client.set(key, json_data)
            
            # 매트릭스는 별도로 바이너리 저장 (효율성)
            matrix_key = f"race_matrix:{race.id}"
            matrix_data = pickle.dumps(race.power_matrix)
            self.binary_client.set(matrix_key, matrix_data)
            
            # 종족 목록에 추가
            self.redis_client.sadd("races", race.id)
            
            return True
        except Exception as e:
            print(f"Error saving race {race.id}: {e}")
            return False
    
    def load_race(self, race_id: str) -> Optional[RaceData]:
        """종족 데이터 로드"""
        try:
            key = f"race:{race_id}"
            json_data = self.redis_client.get(key)
            
            if not json_data:
                return None
            
            data = json.loads(json_data)
            
            # 매트릭스 별도 로드
            matrix_key = f"race_matrix:{race_id}"
            matrix_data = self.binary_client.get(matrix_key)
            if matrix_data:
                data['power_matrix'] = pickle.loads(matrix_data).tolist()
            
            return RaceData.from_dict(data)
        except Exception as e:
            print(f"Error loading race {race_id}: {e}")
            return None
    
    def get_all_races(self) -> List[RaceData]:
        """모든 종족 데이터 로드"""
        race_ids = self.redis_client.smembers("races")
        races = []
        
        for race_id in race_ids:
            race = self.load_race(race_id)
            if race:
                races.append(race)
        
        return races
    
    # 타워 데이터 관리
    def save_tower(self, tower: TowerData) -> bool:
        """타워 데이터 저장"""
        try:
            key = f"tower:{tower.id}"
            data = tower.to_dict()
            
            json_data = json.dumps(data, ensure_ascii=False)
            self.redis_client.set(key, json_data)
            
            # 종족별 타워 목록에 추가
            race_towers_key = f"race_towers:{tower.race_id}"
            self.redis_client.sadd(race_towers_key, tower.id)
            
            # 티어별 타워 목록에 추가
            tier_towers_key = f"tier_towers:{tower.tier}"
            self.redis_client.sadd(tier_towers_key, tower.id)
            
            return True
        except Exception as e:
            print(f"Error saving tower {tower.id}: {e}")
            return False
    
    def load_tower(self, tower_id: str) -> Optional[TowerData]:
        """타워 데이터 로드"""
        try:
            key = f"tower:{tower_id}"
            json_data = self.redis_client.get(key)
            
            if not json_data:
                return None
            
            data = json.loads(json_data)
            return TowerData.from_dict(data)
        except Exception as e:
            print(f"Error loading tower {tower_id}: {e}")
            return None
    
    def get_race_towers(self, race_id: str) -> List[TowerData]:
        """특정 종족의 모든 타워 로드"""
        race_towers_key = f"race_towers:{race_id}"
        tower_ids = self.redis_client.smembers(race_towers_key)
        towers = []
        
        for tower_id in tower_ids:
            tower = self.load_tower(tower_id)
            if tower:
                towers.append(tower)
        
        return towers
    
    # 게임 상태 관리
    def save_game_state(self, game_id: str, state: Dict[str, Any]) -> bool:
        """게임 상태 저장"""
        try:
            key = f"game_state:{game_id}"
            
            # NumPy 배열을 리스트로 변환
            serializable_state = self._make_serializable(state)
            
            json_data = json.dumps(serializable_state, ensure_ascii=False)
            self.redis_client.set(key, json_data)
            
            # TTL 설정 (24시간)
            self.redis_client.expire(key, 86400)
            
            return True
        except Exception as e:
            print(f"Error saving game state {game_id}: {e}")
            return False
    
    def load_game_state(self, game_id: str) -> Optional[Dict[str, Any]]:
        """게임 상태 로드"""
        try:
            key = f"game_state:{game_id}"
            json_data = self.redis_client.get(key)
            
            if not json_data:
                return None
            
            state = json.loads(json_data)
            return self._restore_numpy_arrays(state)
        except Exception as e:
            print(f"Error loading game state {game_id}: {e}")
            return None
    
    # 밸런스 히스토리 관리
    def save_balance_metrics(self, game_id: str, timestamp: float, metrics: Dict) -> bool:
        """밸런스 메트릭 저장"""
        try:
            key = f"balance_history:{game_id}"
            
            # 시계열 데이터로 저장
            self.redis_client.zadd(key, {json.dumps(metrics): timestamp})
            
            # 최근 1시간 데이터만 유지
            cutoff_time = timestamp - 3600
            self.redis_client.zremrangebyscore(key, 0, cutoff_time)
            
            return True
        except Exception as e:
            print(f"Error saving balance metrics for {game_id}: {e}")
            return False
    
    def get_balance_history(self, game_id: str, start_time: float = 0) -> List[Dict]:
        """밸런스 히스토리 조회"""
        try:
            key = f"balance_history:{game_id}"
            
            # 시간 범위로 조회
            data = self.redis_client.zrangebyscore(key, start_time, '+inf', withscores=True)
            
            history = []
            for metrics_json, timestamp in data:
                metrics = json.loads(metrics_json)
                metrics['timestamp'] = timestamp
                history.append(metrics)
            
            return history
        except Exception as e:
            print(f"Error loading balance history for {game_id}: {e}")
            return []
    
    # 유틸리티 메서드
    def _make_serializable(self, obj: Any) -> Any:
        """NumPy 배열을 JSON 직렬화 가능한 형태로 변환"""
        if isinstance(obj, np.ndarray):
            return obj.tolist()
        elif isinstance(obj, dict):
            return {k: self._make_serializable(v) for k, v in obj.items()}
        elif isinstance(obj, list):
            return [self._make_serializable(item) for item in obj]
        else:
            return obj
    
    def _restore_numpy_arrays(self, obj: Any) -> Any:
        """리스트를 NumPy 배열로 복원 (필요한 경우)"""
        # 이 메서드는 특정 키 패턴에 따라 리스트를 NumPy 배열로 변환
        if isinstance(obj, dict):
            result = {}
            for k, v in obj.items():
                if k.endswith('_matrix') and isinstance(v, list):
                    result[k] = np.array(v)
                else:
                    result[k] = self._restore_numpy_arrays(v)
            return result
        elif isinstance(obj, list):
            return [self._restore_numpy_arrays(item) for item in obj]
        else:
            return obj
    
    # 캐시 관리
    def clear_cache(self, pattern: str = None) -> int:
        """캐시 정리"""
        if pattern:
            keys = self.redis_client.keys(pattern)
            if keys:
                return self.redis_client.delete(*keys)
        return 0
    
    def get_cache_stats(self) -> Dict[str, int]:
        """캐시 통계"""
        info = self.redis_client.info()
        return {
            'used_memory': info.get('used_memory', 0),
            'connected_clients': info.get('connected_clients', 0),
            'total_commands_processed': info.get('total_commands_processed', 0),
            'keyspace_hits': info.get('keyspace_hits', 0),
            'keyspace_misses': info.get('keyspace_misses', 0)
        }

# 사용 예시
if __name__ == "__main__":
    # Redis 저장소 초기화
    store = DefenseAlliesRedisStore()
    
    # 종족 데이터 생성 및 저장 예시
    human_race = RaceData(
        id="human_alliance",
        name="Human Alliance",
        theme="균형과 적응성",
        power_matrix=np.array([[1.0, 1.0], [1.0, 1.0]]),
        matrix_properties=MatrixProperties(2.0, 0.0, 2.0, [2.0, 0.0]),
        towers=TowerSet(
            basic=["knight_fortress", "merchant_guild", "mage_tower"],
            advanced=["castle_walls", "cathedral", "royal_palace"],
            cooperation=["alliance_fortress", "peace_tower", "unity_command"]
        ),
        environment_modifiers=EnvironmentModifiers()
    )
    
    # 저장
    success = store.save_race(human_race)
    print(f"Human race saved: {success}")
    
    # 로드
    loaded_race = store.load_race("human_alliance")
    print(f"Loaded race: {loaded_race.name if loaded_race else 'None'}")
```

---

**다음 단계**: 실제 게임 서버 통합 및 성능 최적화
