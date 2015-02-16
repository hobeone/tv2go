episodeService = angular.module('tv2go.episodesService', ['ngResource']);

episodeService.factory('Episode', ['$cacheFactory','$resource',
  function($cacheFactory, $resource) {
    return $resource('http://localhost:9001/api/1/shows/:showid/episodes/:episodeid', {}, {
      all: {
        method: "GET",
        cache: true,
        isArray: true,
      },
      update: {
        method: "PUT"
      },
    });
  }]);

angular.module('tv2go.models.episodes',['tv2go.episodesService'])
.service('EpisodesModel', function($http, $q, Episode){
  var model = this;
  var episodes;

  function cacheEpisodes(result) {
    episodes = result;
    return episodes;
  }

  model.getEpisodes = function(showid) {
    var deferred = $q.defer();

      Episode.all(
        {
          showid: showid,
        }
      ).$promise.then(function(episodes){
        deferred.resolve(cacheEpisodes(episodes));
      });
    return deferred.promise;
  };

  model.createEpisode = function(episode) {
    episode.$save();
    episodes.push(episode);
  };

  model.updateEpisode = function(episode) {
    console.log(episode);
    episode.$update({
      showid: episode.showid,
    });

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
