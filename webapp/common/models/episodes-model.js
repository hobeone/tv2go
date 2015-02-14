episodeService = angular.module('tv2go.episodesService', ['ngResource']);

episodeService.factory('Episode', ['$resource',
  function($resource) {
    return $resource('data/:episodeId.json', {}, {
      update: {
        method: "PUT"
      },
      query: {
        method: "GET",
        params: {
          episodeId: 'episodes',
        },
        isArray:true
      }
    });
  }]);

angular.module('tv2go.models.episodes',['tv2go.episodesService'])
.service('EpisodesModel', function($http, $q, Episode){
  var model = this,
    episodes;

  function cacheEpisodes(result) {
    episodes = result;
    return episodes;
  }

  model.getEpisodes = function() {
    var deferred = $q.defer();

    if(episodes) {
      deferred.resolve(episodes);
    } else {
      Episode.query().$promise.then(function(episodes){
        deferred.resolve(cacheEpisodes(episodes));
      });
    }

    return deferred.promise;
  };

  model.createEpisode = function(episode) {
    console.log(episode);
    episode.$save();
    episodes.push(episode);
  };

  model.updateEpisode = function(episode) {
    episode.$save();

    var index = _.findIndex(episodes, function(e){
      return e.id == episode.id;
    });
    episodes[index] = episode;
  };

  model.findEpisode = function(episodeId) {
    return _.find(episodes, function(episode) {
      return episode.id === parseInt(episodeId, 10);
    });
  };

  model.getEpisodeById = function(episodeId) {
    var deferred = $q.defer();

    episode.$save();
    if(episodes) {
      deferred.resolve(model.findEpisode(episodeId));
    } else {
      model.getEpisodes().then(function() {
        deferred.resolve(model.findEpisode(episodeId));
      });
    }

    return deferred.promise;
  };


  model.deleteEpisode = function(episode) {
    episode.$delete({episodeId: episode.id});
    _.remove(episodes, function(e) {
      return e.id == episode.id;
    });
  };
})
;
